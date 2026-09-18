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
	cycleCount         atomic.Int32
}

// DiscoveryStats holds stats for the proxy status API
type DiscoveryStats struct {
	IsRunning          bool      `json:"is_running"`
	LastDiscoveryTime  time.Time `json:"last_discovery_time"`
	LastDiscoveryCount int       `json:"last_discovery_count"`
	TotalDiscovered    int       `json:"total_discovered"`
	TotalHealthy       int       `json:"total_healthy"`
	CycleCount         int       `json:"cycle_count"`
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

	// Run discovery engine
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
		CycleCount:         int(GlobalDiscovery.cycleCount.Load()),
	}
}

func (d *ProxyDiscovery) run() {
	logger.Log.Info("🔍 ProxyDiscovery: Starting Indian proxy auto-discovery engine (10 sources)")

	// Initial aggressive discovery
	d.discoverAndValidate()

	// First hour: discover every 5 minutes (aggressive warm-up)
	// After first hour: use configured interval
	fastTicker := time.NewTicker(5 * time.Minute)
	healthTicker := time.NewTicker(d.healthCheckInterval)
	normalTimer := time.NewTimer(1 * time.Hour)

	defer fastTicker.Stop()
	defer healthTicker.Stop()
	defer normalTimer.Stop()

	for {
		select {
		case <-fastTicker.C:
			cycles := d.cycleCount.Load()
			if cycles < 12 { // First hour (12 * 5min = 60min)
				d.discoverAndValidate()
			}
		case <-normalTimer.C:
			// Switch to normal interval after 1 hour
			fastTicker.Stop()
			normalTicker := time.NewTicker(d.discoveryInterval)
			defer normalTicker.Stop()
			go func() {
				for {
					select {
					case <-normalTicker.C:
						d.discoverAndValidate()
					case <-d.ctx.Done():
						return
					}
				}
			}()
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

	cycle := d.cycleCount.Add(1)
	logger.Log.Infof("🔍 ProxyDiscovery: Starting discovery cycle #%d...", cycle)

	// Collect proxies from ALL sources concurrently
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
		{"ProxyListDownload", d.fetchFromProxyListDownload},
		{"SpysOne", d.fetchFromSpysOne},
		{"PubProxy", d.fetchFromPubProxy},
		{"FreeProxyCZ", d.fetchFromFreeProxyCZ},
		{"ProxyNova", d.fetchFromProxyNova},
		{"HideMyLife", d.fetchFromHideMy},
		{"OpenProxyList", d.fetchFromOpenProxyList},
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
				logger.Log.Infof("🔍 [%s] → %d Indian proxies", name, len(proxies))
			} else {
				logger.Log.Warnf("🔍 [%s] → 0 proxies", name)
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

	logger.Log.Infof("🔍 ProxyDiscovery: Found %d unique Indian proxies, starting GMS validation (40 workers)...", len(unique))

	if len(unique) == 0 {
		return
	}

	// Validate proxies against GMS portal (40 concurrent validators)
	validated := d.validateProxies(unique, 40)

	d.totalHealthy.Store(int32(len(validated)))

	if len(validated) > 0 {
		d.pool.AddProxies(validated)
		logger.Log.Infof("✅ ProxyDiscovery: Added %d GMS-validated Indian proxies to pool", len(validated))

		// Flush to Supabase cache
		if GlobalProxyCache != nil {
			GlobalProxyCache.FlushTopProxies(d.pool, 20)
		}
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
	top := d.pool.getBestProxies(minInt(15, count))
	var urls []string
	for _, p := range top {
		urls = append(urls, p.URL)
	}

	validated := d.validateProxies(urls, 15)
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
		Timeout:   12 * time.Second,
	}

	ctx, cancel := context.WithTimeout(d.ctx, 10*time.Second)
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

// ==========================================
// 10 PROXY SOURCES — All filtered for India
// ==========================================

// 1. ProxyScrape — HTTP + SOCKS5
func (d *ProxyDiscovery) fetchFromProxyScrape(ctx context.Context) []string {
	var allProxies []string

	urls := []string{
		"https://api.proxyscrape.com/v2/?request=displayproxies&protocol=http&timeout=10000&country=IN&ssl=all&anonymity=all",
		"https://api.proxyscrape.com/v2/?request=displayproxies&protocol=socks5&timeout=10000&country=IN&ssl=all&anonymity=all",
		"https://api.proxyscrape.com/v2/?request=displayproxies&protocol=socks4&timeout=10000&country=IN&ssl=all&anonymity=all",
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
				if strings.Contains(apiURL, "socks5") {
					allProxies = append(allProxies, "socks5://"+proxy)
				} else if strings.Contains(apiURL, "socks4") {
					allProxies = append(allProxies, "socks4://"+proxy)
				} else {
					allProxies = append(allProxies, "http://"+proxy)
				}
			}
		}
	}

	return allProxies
}

// 2. GeoNode — JSON API
func (d *ProxyDiscovery) fetchFromGeoNode(ctx context.Context) []string {
	var allProxies []string

	pages := []string{
		"https://proxylist.geonode.com/api/proxy-list?limit=100&page=1&sort_by=lastChecked&sort_type=desc&country=IN&filterUpTime=90&speed=fast",
		"https://proxylist.geonode.com/api/proxy-list?limit=100&page=2&sort_by=lastChecked&sort_type=desc&country=IN&filterUpTime=90&speed=fast",
	}

	client := &http.Client{Timeout: 20 * time.Second}

	for _, apiURL := range pages {
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

		var result struct {
			Data []struct {
				IP        string   `json:"ip"`
				Port      string   `json:"port"`
				Protocols []string `json:"protocols"`
			} `json:"data"`
		}

		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			continue
		}

		for _, entry := range result.Data {
			if entry.IP == "" || entry.Port == "" {
				continue
			}
			protocol := "http"
			for _, p := range entry.Protocols {
				if strings.ToLower(p) == "socks5" {
					protocol = "socks5"
					break
				}
			}
			allProxies = append(allProxies, fmt.Sprintf("%s://%s:%s", protocol, entry.IP, entry.Port))
		}
	}

	return allProxies
}

// 3. FreeProxyList.net — HTML scrape
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

	ipPortRegex := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\s*</td>\s*<td>\s*(\d{2,5})`)
	countryRegex := regexp.MustCompile(`<td>\s*IN\s*</td>`)

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

// 4. ProxyList.download — plain text lists
func (d *ProxyDiscovery) fetchFromProxyListDownload(ctx context.Context) []string {
	var allProxies []string

	urls := []string{
		"https://www.proxy-list.download/api/v1/get?type=http&country=IN",
		"https://www.proxy-list.download/api/v1/get?type=https&country=IN",
		"https://www.proxy-list.download/api/v1/get?type=socks5&country=IN",
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

// 5. Spys.one — India-filtered proxy list
func (d *ProxyDiscovery) fetchFromSpysOne(ctx context.Context) []string {
	var allProxies []string

	client := &http.Client{Timeout: 20 * time.Second}

	// Spys.one provides a text format for Indian proxies
	req, err := http.NewRequestWithContext(ctx, "GET", "https://spys.me/proxy.txt", nil)
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

	// Format: IP:PORT CC-ANON-S/H
	// Filter for lines containing " IN-" (India country code)
	lines := strings.Split(string(bodyBytes), "\n")
	ipPortRegex := regexp.MustCompile(`^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d{2,5})\s+IN-`)

	for _, line := range lines {
		matches := ipPortRegex.FindStringSubmatch(strings.TrimSpace(line))
		if len(matches) >= 2 {
			allProxies = append(allProxies, "http://"+matches[1])
		}
	}

	return allProxies
}

// 6. PubProxy — JSON API with India filter
func (d *ProxyDiscovery) fetchFromPubProxy(ctx context.Context) []string {
	var allProxies []string

	client := &http.Client{Timeout: 20 * time.Second}

	// PubProxy has rate limits, so we make a few requests
	for i := 0; i < 5; i++ {
		if ctx.Err() != nil {
			break
		}

		apiURL := "http://pubproxy.com/api/proxy?country=IN&type=http&limit=5&format=json"
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

		var result struct {
			Data []struct {
				IP   string `json:"ip"`
				Port string `json:"port"`
				Type string `json:"type"`
			} `json:"data"`
		}

		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			continue
		}

		for _, entry := range result.Data {
			protocol := "http"
			if strings.ToLower(entry.Type) == "socks5" {
				protocol = "socks5"
			}
			allProxies = append(allProxies, fmt.Sprintf("%s://%s:%s", protocol, entry.IP, entry.Port))
		}

		// Small delay to avoid rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	return allProxies
}

// 7. FreeProxyCZ — Czech free proxy site with India filter
func (d *ProxyDiscovery) fetchFromFreeProxyCZ(ctx context.Context) []string {
	var allProxies []string

	client := &http.Client{Timeout: 20 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "GET", "http://free-proxy.cz/en/proxylist/country/IN/all/ping/all", nil)
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

	// Extract IP:PORT from the page — they use base64 encoded IPs in script tags
	// Fallback: try direct IP:PORT regex
	ipPortRegex := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\s*</td>\s*<td[^>]*>\s*(\d{2,5})`)
	matches := ipPortRegex.FindAllStringSubmatch(body, -1)
	for _, m := range matches {
		if len(m) >= 3 {
			allProxies = append(allProxies, fmt.Sprintf("http://%s:%s", m[1], m[2]))
		}
	}

	return allProxies
}

// 8. ProxyNova — India page
func (d *ProxyDiscovery) fetchFromProxyNova(ctx context.Context) []string {
	var allProxies []string

	client := &http.Client{Timeout: 20 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.proxynova.com/proxy-server-list/country-in/", nil)
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

	// ProxyNova puts IP inside a <abbr> tag with JS obfuscation
	// Try direct regex fallback
	ipRegex := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	portRegex := regexp.MustCompile(`<td[^>]*>\s*(\d{2,5})\s*</td>`)

	rows := strings.Split(body, "<tr")
	for _, row := range rows {
		ips := ipRegex.FindStringSubmatch(row)
		ports := portRegex.FindStringSubmatch(row)
		if len(ips) >= 2 && len(ports) >= 2 {
			allProxies = append(allProxies, fmt.Sprintf("http://%s:%s", ips[1], ports[1]))
		}
	}

	return allProxies
}

// 9. HideMy.life — API-based
func (d *ProxyDiscovery) fetchFromHideMy(ctx context.Context) []string {
	var allProxies []string

	client := &http.Client{Timeout: 20 * time.Second}

	// hidemy.life has a public API
	req, err := http.NewRequestWithContext(ctx, "GET", "https://hidemy.life/api/proxylist.php?out=plain&country=IN&maxtime=5000", nil)
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

	lines := strings.Split(string(bodyBytes), "\n")
	for _, line := range lines {
		proxy := strings.TrimSpace(line)
		if proxy != "" && strings.Contains(proxy, ":") {
			allProxies = append(allProxies, "http://"+proxy)
		}
	}

	return allProxies
}

// 10. OpenProxyList — aggregator text list
func (d *ProxyDiscovery) fetchFromOpenProxyList(ctx context.Context) []string {
	var allProxies []string

	urls := []string{
		"https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/http.txt",
		"https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/socks5.txt",
		"https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/http.txt",
		"https://raw.githubusercontent.com/monosans/proxy-list/main/proxies_geolocation/http.txt",
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

		// For geolocation file, lines look like: IP:PORT|IN|... 
		// For plain files, we take all and let GMS validation filter non-Indian
		isGeo := strings.Contains(apiURL, "geolocation")

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || !strings.Contains(line, ":") {
				continue
			}

			if isGeo {
				// Format: IP:PORT|CountryCode|...
				parts := strings.Split(line, "|")
				if len(parts) >= 2 && strings.TrimSpace(parts[1]) == "IN" {
					proxy := strings.TrimSpace(parts[0])
					if strings.Contains(apiURL, "socks5") {
						allProxies = append(allProxies, "socks5://"+proxy)
					} else {
						allProxies = append(allProxies, "http://"+proxy)
					}
				}
			} else {
				// Plain list — we can't filter by country, so just take all
				// GMS validation will filter out non-working ones
				// Only take first 200 to avoid overwhelming validation
				if len(allProxies) >= 200 {
					break
				}
				proxy := line
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

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
