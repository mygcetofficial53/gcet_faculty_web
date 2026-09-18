package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"gcet-web-backend/internal/logger"
)

// ProxyCache handles persisting working proxies to Supabase so they survive Render restarts.
type ProxyCache struct {
	supabaseURL string
	supabaseKey string
	mu          sync.RWMutex
	enabled     bool
}

// CachedProxy represents a proxy row in the proxy_cache table
type CachedProxy struct {
	URL           string    `json:"url"`
	Successes     int32     `json:"successes"`
	Failures      int32     `json:"failures"`
	AvgMs         int64     `json:"avg_ms"`
	LastValidated time.Time `json:"last_validated"`
	CreatedAt     time.Time `json:"created_at"`
}

// GlobalProxyCache is the singleton
var GlobalProxyCache *ProxyCache

// InitProxyCache creates the Supabase-backed proxy cache
func InitProxyCache(supabaseURL, supabaseKey string) {
	if supabaseURL == "" || supabaseKey == "" {
		logger.Log.Warn("ProxyCache: Supabase credentials missing — persistence disabled")
		GlobalProxyCache = &ProxyCache{enabled: false}
		return
	}

	GlobalProxyCache = &ProxyCache{
		supabaseURL: strings.TrimRight(supabaseURL, "/"),
		supabaseKey: supabaseKey,
		enabled:     true,
	}

	// Ensure the proxy_cache table exists (create via RPC or direct insert)
	GlobalProxyCache.ensureTable()

	logger.Log.Info("💾 ProxyCache: Supabase persistence initialized")
}

// ensureTable creates the proxy_cache table if it doesn't exist, using Supabase REST API
func (pc *ProxyCache) ensureTable() {
	if !pc.enabled {
		return
	}
	// We'll try a simple SELECT. If the table doesn't exist, we log a warning.
	// The user should create the table in the Supabase dashboard.
	url := pc.supabaseURL + "/rest/v1/proxy_cache?select=url&limit=1"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", pc.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+pc.supabaseKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Warnf("ProxyCache: Could not reach Supabase: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 || resp.StatusCode == 400 {
		logger.Log.Warn("ProxyCache: proxy_cache table does not exist in Supabase. " +
			"Please create it with: CREATE TABLE proxy_cache (" +
			"url TEXT PRIMARY KEY, successes INT DEFAULT 0, failures INT DEFAULT 0, " +
			"avg_ms INT DEFAULT 0, last_validated TIMESTAMPTZ DEFAULT NOW(), " +
			"created_at TIMESTAMPTZ DEFAULT NOW());")
		pc.enabled = false
	} else {
		logger.Log.Info("💾 ProxyCache: proxy_cache table verified")
	}
}

// LoadCachedProxies fetches all cached proxies from Supabase and returns them as ProxyEntries.
// Only returns proxies validated within the last 6 hours.
func (pc *ProxyCache) LoadCachedProxies() []*ProxyEntry {
	if !pc.enabled {
		return nil
	}

	pc.mu.RLock()
	defer pc.mu.RUnlock()

	// Fetch proxies validated within the last 6 hours, ordered by success rate
	cutoff := time.Now().UTC().Add(-6 * time.Hour).Format(time.RFC3339)
	url := fmt.Sprintf("%s/rest/v1/proxy_cache?select=*&last_validated=gte.%s&order=successes.desc&limit=50",
		pc.supabaseURL, cutoff)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Log.Errorf("ProxyCache: Failed to create request: %v", err)
		return nil
	}
	req.Header.Set("apikey", pc.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+pc.supabaseKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Errorf("ProxyCache: Failed to fetch cached proxies: %v", err)
		return nil
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var cached []CachedProxy
	if err := json.Unmarshal(bodyBytes, &cached); err != nil {
		logger.Log.Errorf("ProxyCache: Failed to parse cached proxies: %v", err)
		return nil
	}

	var entries []*ProxyEntry
	for _, cp := range cached {
		entries = append(entries, &ProxyEntry{
			URL:       cp.URL,
			Successes: cp.Successes,
			Failures:  cp.Failures,
			AvgMs:     cp.AvgMs,
		})
	}

	if len(entries) > 0 {
		logger.Log.Infof("💾 ProxyCache: Loaded %d cached proxies from Supabase", len(entries))
	}
	return entries
}

// FlushTopProxies saves the top N working proxies to Supabase
func (pc *ProxyCache) FlushTopProxies(pool *ProxyPool, maxEntries int) {
	if !pc.enabled || pool == nil {
		return
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	top := pool.getBestProxies(maxEntries)
	if len(top) == 0 {
		return
	}

	// Build upsert payload
	type upsertRow struct {
		URL           string `json:"url"`
		Successes     int32  `json:"successes"`
		Failures      int32  `json:"failures"`
		AvgMs         int64  `json:"avg_ms"`
		LastValidated string `json:"last_validated"`
	}

	var rows []upsertRow
	now := time.Now().UTC().Format(time.RFC3339)
	for _, pe := range top {
		// Only cache proxies that have at least 1 success
		if pe.Successes > 0 {
			rows = append(rows, upsertRow{
				URL:           pe.URL,
				Successes:     pe.Successes,
				Failures:      pe.Failures,
				AvgMs:         pe.AvgMs,
				LastValidated: now,
			})
		}
	}

	if len(rows) == 0 {
		return
	}

	payload, err := json.Marshal(rows)
	if err != nil {
		logger.Log.Errorf("ProxyCache: Failed to marshal flush payload: %v", err)
		return
	}

	url := pc.supabaseURL + "/rest/v1/proxy_cache"
	req, err := http.NewRequest("POST", url, strings.NewReader(string(payload)))
	if err != nil {
		logger.Log.Errorf("ProxyCache: Failed to create flush request: %v", err)
		return
	}

	req.Header.Set("apikey", pc.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+pc.supabaseKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "resolution=merge-duplicates") // Upsert on conflict

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Errorf("ProxyCache: Failed to flush to Supabase: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logger.Log.Infof("💾 ProxyCache: Flushed %d proxies to Supabase", len(rows))
	} else {
		body, _ := io.ReadAll(resp.Body)
		logger.Log.Errorf("ProxyCache: Flush failed (HTTP %d): %s", resp.StatusCode, string(body))
	}
}

// PurgeStale removes cached proxies older than maxAge from Supabase
func (pc *ProxyCache) PurgeStale(maxAge time.Duration) {
	if !pc.enabled {
		return
	}

	cutoff := time.Now().UTC().Add(-maxAge).Format(time.RFC3339)
	url := fmt.Sprintf("%s/rest/v1/proxy_cache?last_validated=lt.%s", pc.supabaseURL, cutoff)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("apikey", pc.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+pc.supabaseKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Errorf("ProxyCache: Failed to purge stale proxies: %v", err)
		return
	}
	defer resp.Body.Close()
	logger.Log.Info("💾 ProxyCache: Purged stale proxies")
}

// StartPeriodicFlush runs a background goroutine that flushes top proxies every interval
func (pc *ProxyCache) StartPeriodicFlush(pool *ProxyPool, flushInterval time.Duration) {
	if !pc.enabled {
		return
	}

	go func() {
		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()

		// Also purge stale entries every flush cycle
		purgeTicker := time.NewTicker(30 * time.Minute)
		defer purgeTicker.Stop()

		for {
			select {
			case <-ticker.C:
				pc.FlushTopProxies(pool, 20)
			case <-purgeTicker.C:
				pc.PurgeStale(6 * time.Hour)
			}
		}
	}()

	logger.Log.Infof("💾 ProxyCache: Periodic flush started (every %v)", flushInterval)
}
