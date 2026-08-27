package service

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gcet-web-backend/internal/logger"
)

// ProxyDiscovery auto-discovers free Indian proxies and validates them against the GMS portal.
type ProxyDiscovery struct {
	gmsURL              string // e.g. "http://202.129.240.148:8080/GIS"
	pool                *ProxyPool
	discoveryInterval   time.Duration
	healthCheckInterval time.Duration
	ctx                 context.Context
	cancel              context.CancelFunc

	// Stats (thread-safe)
	lastDiscoveryTime  atomic.Value // time.Time
	lastDiscoveryCount atomic.Int32
	totalDiscovered    atomic.Int32
	totalHealthy       atomic.Int32
	isRunning          atomic.Bool
}

// DiscoveryStats holds stats for the proxy status API
type DiscoveryStats struct {
	IsRunning          bool      `json:"is_running"`
	LastDiscoveryTime  time.Time `json:"last_discovery_time"`
	LastDiscoveryCount int       `json:"last_discovery_count"`
	TotalDiscovered    int       `json:"total_discovered"`
	TotalHealthy       int       `json:"total_healthy"`
}

// GlobalDiscovery is the singleton instance
var GlobalDiscovery *ProxyDiscovery

// StartProxyDiscovery initializes and starts the auto-discovery engine
func StartProxyDiscovery(gmsURL string, pool *ProxyPool, discoveryInterval, healthCheckInterval time.Duration) {
	if pool == nil {
		logger.Log.Warn("ProxyDiscovery: No proxy pool — creating standalone pool")
		ctx, cancel := context.WithCancel(context.Background())
		pool = &ProxyPool{
			ctx:    ctx,
			cancel: cancel,
		}
		GlobalProxyPool = pool
	}

	ctx, cancel := context.WithCancel(context.Background())
	GlobalDiscovery = &ProxyDiscovery{
		gmsURL:              gmsURL,
		pool:                pool,
		discoveryInterval:   discoveryInterval,
		healthCheckInterval: healthCheckInterval,
		ctx:                 ctx,
		cancel:              cancel,
	}

	GlobalDiscovery.lastDiscoveryTime.Store(time.Time{})

	// Run initial discovery immediately
	go GlobalDiscovery.run()
}

// StopProxyDiscovery gracefully stops the discovery engine
func StopProxyDiscovery() {
	if GlobalDiscovery != nil && GlobalDiscovery.cancel != nil {
		GlobalDiscovery.cancel()
	}
}

// GetDiscoveryStats returns current discovery stats
func GetDiscoveryStats() *DiscoveryStats {
	if GlobalDiscovery == nil {
		return &DiscoveryStats{}
	}

	lastTime, _ := GlobalDiscovery.lastDiscoveryTime.Load().(time.Time)

	return &DiscoveryStats{
		IsRunning:          GlobalDiscovery.isRunning.Load(),
		LastDiscoveryTime:  lastTime,
		LastDiscoveryCount: int(GlobalDiscovery.lastDiscoveryCount.Load()),
		TotalDiscovered:    int(GlobalDiscovery.totalDiscovered.Load()),
		TotalHealthy:       int(GlobalDiscovery.totalHealthy.Load()),
	}
}

func (d *ProxyDiscovery) run() {
	logger.Log.Info("🔍 ProxyDiscovery: Starting Indian proxy auto-discovery engine")

	// Initial discovery
	d.discoverAndValidate()

	// Periodic discovery
	discoveryTicker := time.NewTicker(d.discoveryInterval)
	// Periodic health check of existing proxies
	healthTicker := time.NewTicker(d.healthCheckInterval)
	defer discoveryTicker.Stop()
	defer healthTicker.Stop()

	for {
		select {
		case <-discoveryTicker.C:
			d.discoverAndValidate()
		case <-healthTicker.C:
			d.healthCheckExisting()
		case <-d.ctx.Done():
			logger.Log.Info("ProxyDiscovery: Shutting down")
			return
		}
	}
}

// discoverAndValidate scrapes proxy sources, filters for Indian proxies, and validates them
func (d *ProxyDiscovery) discoverAndValidate() {
	d.isRunning.Store(true)
	defer d.isRunning.Store(false)

	logger.Log.Info("🔍 ProxyDiscovery: Starting discovery cycle...")

	// Collect proxies from all sources
	var allProxies []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	sources := []struct {
		name string
		fn   func(context.Context) []string
	}{
		{"ProxyScrape", d.fetchFromProxyScrape},
		{"GeoNode", d.fetchFromGeoNode},
		{"FreeProxyList", d.fetchFromFreeProxyList},
	}

	for _, src := range sources {
		wg.Add(1)
		go func(name string, fn func(context.Context) []string) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(d.ctx, 30*time.Second)
			defer cancel()

			proxies := fn(ctx)
			if len(proxies) > 0 {
				mu.Lock()
				allProxies = append(allProxies, proxies...)
				mu.Unlock()
				logger.Log.Infof("🔍 ProxyDiscovery: %s returned %d Indian proxies", name, len(proxies))
			} else {
				logger.Log.Warnf("🔍 ProxyDiscovery: %s returned 0 proxies", name)
			}
		}(src.name, src.fn)
	}

	wg.Wait()

	// Deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, p := range allProxies {
		p = strings.TrimSpace(p)
		if p != "" && !seen[p] {
			seen[p] = true
			unique = append(unique, p)
		}
	}

	d.totalDiscovered.Store(int32(len(unique)))
	d.lastDiscoveryTime.Store(time.Now())
	d.lastDiscoveryCount.Store(int32(len(unique)))

	logger.Log.Infof("🔍 ProxyDiscovery: Found %d unique Indian proxies, starting GMS validation...", len(unique))

	if len(unique) == 0 {
		return
	}

	// Validate proxies against GMS portal (concurrent, limited to 20 goroutines)
	validated := d.validateProxies(unique, 20)

	d.totalHealthy.Store(int32(len(validated)))

	if len(validated) > 0 {
		d.pool.AddProxies(validated)
		logger.Log.Infof("✅ ProxyDiscovery: Added %d GMS-validated Indian proxies to pool", len(validated))
	} else {
		logger.Log.Warn("⚠️ ProxyDiscovery: No proxies passed GMS validation this cycle")
	}
}

// healthCheckExisting re-validates the top proxies in the pool
func (d *ProxyDiscovery) healthCheckExisting() {
	if d.pool == nil {
		return
	}

	d.pool.mu.RLock()
	count := len(d.pool.proxies)
	d.pool.mu.RUnlock()

	if count == 0 {
		return
	}

	logger.Log.Infof("🏥 ProxyDiscovery: Health-checking %d proxies in pool...", count)

	// Remove proxies with too many failures
	d.pool.RemoveDeadProxies(10)

	// Re-validate the top N proxies
	top := d.pool.getBestProxies(min(10, count))
	var urls []string
	for _, p := range top {
		urls = append(urls, p.URL)
	}

	validated := d.validateProxies(urls, 10)
	healthyCount := len(validated)

	d.totalHealthy.Store(int32(healthyCount))
	logger.Log.Infof("🏥 ProxyDiscovery: %d/%d proxies are healthy", healthyCount, len(urls))
}

// validateProxies tests a list of proxy URLs against the GMS portal
func (d *ProxyDiscovery) validateProxies(proxyURLs []string, concurrency int) []*ProxyEntry {
	var validated []*ProxyEntry
	var mu sync.Mutex
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	gmsLoginURL := d.gmsURL + "/index.jsp"

	for _, proxyStr := range proxyURLs {
		wg.Add(1)
		sem <- struct{}{} // Acquire slot

		go func(pStr string) {
			defer wg.Done()
			defer func() { <-sem }() // Release slot

			if d.ctx.Err() != nil {
				return
			}

			start := time.Now()
			if d.testProxyAgainstGMS(pStr, gmsLoginURL) {
				dur := time.Since(start)
				mu.Lock()
				validated = append(validated, &ProxyEntry{
					URL:       pStr,
					Successes: 1,
					AvgMs:     dur.Milliseconds(),
				})
				mu.Unlock()
			}
		}(proxyStr)
	}

	wg.Wait()
	return validated
}

// testProxyAgainstGMS checks if a proxy can reach the GMS portal login page
func (d *ProxyDiscovery) testProxyAgainstGMS(proxyStr, gmsLoginURL string) bool {
	// Normalize proxy URL
	if !strings.Contains(proxyStr, "://") {
		proxyStr = "http://" + proxyStr
	}

	proxyURL, err := url.Parse(proxyStr)
	if err != nil {
		return false
	}

	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{
		Proxy:             http.ProxyURL(proxyURL),
		ForceAttemptHTTP2: false,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		TLSNextProto:      make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}

	client := &http.Client{
		Transport: transport,
		Jar:       jar,
		Timeout:   15 * time.Second,
	}

	ctx, cancel := context.WithTimeout(d.ctx, 12*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", gmsLoginURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 50*1024)) // Read max 50KB
	if err != nil {
		return false
	}

	body := strings.ToLower(string(bodyBytes))

	// The GMS portal login page should contain "Faculty Login" or "login_id"
	return strings.Contains(body, "faculty login") || strings.Contains(body, "login_id")
}

// --- Proxy Source Scrapers ---

// fetchFromProxyScrape uses the ProxyScrape API to get Indian HTTP/SOCKS5 proxies
func (d *ProxyDiscovery) fetchFromProxyScrape(ctx context.Context) []string {
	var allProxies []string

	// HTTP proxies
	urls := []string{
		"https://api.proxyscrape.com/v2/?request=displayproxies&protocol=http&timeout=10000&country=IN&ssl=all&anonymity=all",
		"https://api.proxyscrape.com/v2/?request=displayproxies&protocol=socks5&timeout=10000&country=IN&ssl=all&anonymity=all",
	}

	client := &http.Client{Timeout: 20 * time.Second}

	for _, apiURL := range urls {
		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		lines := strings.Split(string(bodyBytes), "\n")
		for _, line := range lines {
			proxy := strings.TrimSpace(line)
			if proxy != "" && strings.Contains(proxy, ":") {
				// Determine protocol from the API URL
				if strings.Contains(apiURL, "socks5") {
					allProxies = append(allProxies, "socks5://"+proxy)
				} else {
					allProxies = append(allProxies, "http://"+proxy)
				}
			}
		}
	}

	return allProxies
}

// fetchFromGeoNode uses the GeoNode free proxy API filtered for India
func (d *ProxyDiscovery) fetchFromGeoNode(ctx context.Context) []string {
	var allProxies []string

	apiURL := "https://proxylist.geonode.com/api/proxy-list?limit=100&page=1&sort_by=lastChecked&sort_type=desc&country=IN&filterUpTime=90&speed=fast"

	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	// Parse JSON response
	var result struct {
		Data []struct {
			IP        string   `json:"ip"`
			Port      string   `json:"port"`
			Protocols []string `json:"protocols"`
		} `json:"data"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil
	}

	for _, entry := range result.Data {
		if entry.IP == "" || entry.Port == "" {
			continue
		}
		// Use the first supported protocol
		protocol := "http"
		for _, p := range entry.Protocols {
			if strings.ToLower(p) == "socks5" {
				protocol = "socks5"
				break
			}
		}
		allProxies = append(allProxies, fmt.Sprintf("%s://%s:%s", protocol, entry.IP, entry.Port))
	}

	return allProxies
}

// fetchFromFreeProxyList scrapes free-proxy-list.net and filters for Indian proxies
func (d *ProxyDiscovery) fetchFromFreeProxyList(ctx context.Context) []string {
	var allProxies []string

	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://free-proxy-list.net/", nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	body := string(bodyBytes)

	// Extract IP:PORT rows — look for lines matching pattern: IP PORT Country
	// The site has a table with IP, PORT, Code, Country, etc.
	// We use a regex to find IP:PORT pairs near "India" or "IN" country code
	ipPortRegex := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\s*</td>\s*<td>\s*(\d{2,5})`)
	countryRegex := regexp.MustCompile(`<td>\s*IN\s*</td>`)

	// Split by table rows
	rows := strings.Split(body, "<tr>")
	for _, row := range rows {
		if !countryRegex.MatchString(row) {
			continue
		}
		matches := ipPortRegex.FindStringSubmatch(row)
		if len(matches) >= 3 {
			ip := matches[1]
			port := matches[2]
			allProxies = append(allProxies, fmt.Sprintf("http://%s:%s", ip, port))
		}
	}

	return allProxies
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
