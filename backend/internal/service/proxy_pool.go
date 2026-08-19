package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gcet-web-backend/internal/logger"
)

// Global Proxy Pool
var GlobalProxyPool *ProxyPool

func init() {
	rand.Seed(time.Now().UnixNano())
}

// ProxyEntry tracks a proxy and its performance
type ProxyEntry struct {
	URL       string
	Successes int32
	Failures  int32
	AvgMs     int64 // average response time in ms
}

type ProxyPool struct {
	url     string
	proxies []*ProxyEntry
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
}

// InitProxyPool initializes the global proxy pool if a URL is provided
func InitProxyPool(listURL string) {
	if listURL == "" {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	GlobalProxyPool = &ProxyPool{
		url:    listURL,
		ctx:    ctx,
		cancel: cancel,
	}

	// Fetch immediately on startup
	GlobalProxyPool.fetchProxies()

	// Fetch periodically every 15 minutes
	go GlobalProxyPool.startRotation()
}

// StopProxyPool stops the background fetcher
func StopProxyPool() {
	if GlobalProxyPool != nil && GlobalProxyPool.cancel != nil {
		GlobalProxyPool.cancel()
	}
}

func (p *ProxyPool) startRotation() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.fetchProxies()
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *ProxyPool) fetchProxies() {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(p.ctx, "GET", p.url, nil)
	if err != nil {
		logger.Log.Errorf("ProxyPool: failed to create request: %v", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Errorf("ProxyPool: failed to fetch proxy list: %v", err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Errorf("ProxyPool: failed to read proxy list: %v", err)
		return
	}

	lines := strings.Split(string(bodyBytes), "\n")
	var entries []*ProxyEntry

	for _, line := range lines {
		proxy := strings.TrimSpace(line)
		if proxy == "" {
			continue
		}

		if !strings.Contains(proxy, "://") {
			proxy = "http://" + proxy
		}

		proxyLower := strings.ToLower(proxy)

		if strings.HasPrefix(proxyLower, "https://") {
			proxy = "http://" + proxy[8:]
			proxyLower = "http://" + proxyLower[8:]
		}

		// Only keep http and socks5 proxies
		if strings.HasPrefix(proxyLower, "http://") || strings.HasPrefix(proxyLower, "socks5://") {
			entries = append(entries, &ProxyEntry{URL: proxy})
		}
	}

	if len(entries) > 0 {
		p.mu.Lock()
		p.proxies = entries
		p.mu.Unlock()
		logger.Log.Infof("ProxyPool: Loaded %d valid proxies", len(entries))
	} else {
		logger.Log.Warn("ProxyPool: No valid proxies found")
	}
}

// getBestProxies returns N proxies, prioritizing ones with high success rates
func (p *ProxyPool) getBestProxies(n int) []*ProxyEntry {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.proxies) == 0 {
		return nil
	}

	// Make a copy to sort
	sorted := make([]*ProxyEntry, len(p.proxies))
	copy(sorted, p.proxies)

	// Sort: proven proxies first (by success count desc, then failures asc)
	sort.Slice(sorted, func(i, j int) bool {
		si := atomic.LoadInt32(&sorted[i].Successes)
		sj := atomic.LoadInt32(&sorted[j].Successes)
		fi := atomic.LoadInt32(&sorted[i].Failures)
		fj := atomic.LoadInt32(&sorted[j].Failures)

		// Proxies with successes and no failures go first
		scoreI := si - fi
		scoreJ := sj - fj
		if scoreI != scoreJ {
			return scoreI > scoreJ
		}
		return si > sj
	})

	// Take top performers + some random ones for discovery
	result := make([]*ProxyEntry, 0, n)

	// Take top proven proxies (up to n/2)
	proven := 0
	for i := 0; i < len(sorted) && proven < n/2; i++ {
		if atomic.LoadInt32(&sorted[i].Successes) > 0 {
			result = append(result, sorted[i])
			proven++
		}
	}

	// Fill remaining slots with random proxies for discovery
	remaining := n - len(result)
	if remaining > 0 {
		// Shuffle the rest
		unproven := sorted[proven:]
		rand.Shuffle(len(unproven), func(i, j int) {
			unproven[i], unproven[j] = unproven[j], unproven[i]
		})
		for i := 0; i < remaining && i < len(unproven); i++ {
			result = append(result, unproven[i])
		}
	}

	return result
}

// GetRandomProxy returns a random proxy URL string (backward compat)
func (p *ProxyPool) GetRandomProxy() string {
	entries := p.getBestProxies(1)
	if len(entries) == 0 {
		return ""
	}
	return entries[0].URL
}

// makeProxyClient creates an http.Client configured to use a specific proxy
func makeProxyClient(proxyStr string, jar *cookiejar.Jar, timeout time.Duration) *http.Client {
	proxyURL, _ := url.Parse(proxyStr)

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = http.ProxyURL(proxyURL)
	transport.ForceAttemptHTTP2 = false
	transport.TLSNextProto = make(map[string]func(authority string, c *tls.Conn) http.RoundTripper)

	return &http.Client{
		Transport: transport,
		Jar:       jar,
		Timeout:   timeout,
	}
}

// cloneCookies creates a new cookie jar and copies cookies for a specific URL
func cloneCookies(src http.CookieJar, targetStr string) http.CookieJar {
	newJar, _ := cookiejar.New(nil)
	if src != nil {
		if parsed, err := url.Parse(targetStr); err == nil {
			newJar.SetCookies(parsed, src.Cookies(parsed))
		}
	}
	return newJar
}

// RaceGet fires a GET request through multiple proxies simultaneously.
// The first successful response wins; all others are cancelled.
func (p *ProxyPool) RaceGet(targetURL string, jar http.CookieJar, numRacers int) (string, error) {
	candidates := p.getBestProxies(numRacers)
	if len(candidates) == 0 {
		return "", fmt.Errorf("no proxies available in pool")
	}

	type raceResult struct {
		body     string
		proxy    *ProxyEntry
		dur      time.Duration
		err      error
		racerJar http.CookieJar
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultCh := make(chan raceResult, len(candidates))

	for _, entry := range candidates {
		go func(pe *ProxyEntry) {
			start := time.Now()
			racerJar := cloneCookies(jar, targetURL)
			client := makeProxyClient(pe.URL, racerJar, 20*time.Second)

			req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
			if err != nil {
				resultCh <- raceResult{err: err, proxy: pe}
				return
			}
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
			req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

			resp, err := client.Do(req)
			if err != nil {
				resultCh <- raceResult{err: err, proxy: pe, dur: time.Since(start)}
				return
			}
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				resultCh <- raceResult{err: err, proxy: pe, dur: time.Since(start)}
				return
			}

			resultCh <- raceResult{body: string(bodyBytes), proxy: pe, dur: time.Since(start), racerJar: racerJar}
		}(entry)
	}

	// Wait for either a success or all failures
	var errors []error
	for i := 0; i < len(candidates); i++ {
		res := <-resultCh
		if res.err == nil {
			// SUCCESS! Record the win and cancel others
			atomic.AddInt32(&res.proxy.Successes, 1)
			atomic.StoreInt64(&res.proxy.AvgMs, res.dur.Milliseconds())
			
			// Merge winner's cookies back to main jar
			if jar != nil {
				if parsed, err := url.Parse(targetURL); err == nil {
					jar.SetCookies(parsed, res.racerJar.Cookies(parsed))
				}
			}

			cancel() // Cancel all other racers
			logger.Log.Infof("ProxyPool RACE: Winner %s responded in %dms", res.proxy.URL, res.dur.Milliseconds())
			return res.body, nil
		}
		// This racer failed
		atomic.AddInt32(&res.proxy.Failures, 1)
		errors = append(errors, fmt.Errorf("proxy %s: %v", res.proxy.URL, res.err))
	}

	return "", fmt.Errorf("all %d proxy racers failed: %v", len(candidates), errors[len(errors)-1])
}

// RacePost fires a POST request through multiple proxies simultaneously.
func (p *ProxyPool) RacePost(targetURL string, contentType string, body string, jar http.CookieJar, numRacers int) (string, error) {
	candidates := p.getBestProxies(numRacers)
	if len(candidates) == 0 {
		return "", fmt.Errorf("no proxies available in pool")
	}

	type raceResult struct {
		body     string
		proxy    *ProxyEntry
		dur      time.Duration
		err      error
		racerJar http.CookieJar
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultCh := make(chan raceResult, len(candidates))

	for _, entry := range candidates {
		go func(pe *ProxyEntry) {
			start := time.Now()
			racerJar := cloneCookies(jar, targetURL)
			client := makeProxyClient(pe.URL, racerJar, 20*time.Second)

			req, err := http.NewRequestWithContext(ctx, "POST", targetURL, strings.NewReader(body))
			if err != nil {
				resultCh <- raceResult{err: err, proxy: pe}
				return
			}
			req.Header.Set("Content-Type", contentType)
			req.Header.Set("Referer", targetURL)
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

			resp, err := client.Do(req)
			if err != nil {
				resultCh <- raceResult{err: err, proxy: pe, dur: time.Since(start)}
				return
			}
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				resultCh <- raceResult{err: err, proxy: pe, dur: time.Since(start)}
				return
			}

			resultCh <- raceResult{body: string(bodyBytes), proxy: pe, dur: time.Since(start), racerJar: racerJar}
		}(entry)
	}

	var errors []error
	for i := 0; i < len(candidates); i++ {
		res := <-resultCh
		if res.err == nil {
			atomic.AddInt32(&res.proxy.Successes, 1)
			atomic.StoreInt64(&res.proxy.AvgMs, res.dur.Milliseconds())
			
			// Merge winner's cookies back to main jar
			if jar != nil {
				if parsed, err := url.Parse(targetURL); err == nil {
					jar.SetCookies(parsed, res.racerJar.Cookies(parsed))
				}
			}

			cancel()
			logger.Log.Infof("ProxyPool RACE: Winner %s responded in %dms", res.proxy.URL, res.dur.Milliseconds())
			return res.body, nil
		}
		atomic.AddInt32(&res.proxy.Failures, 1)
		errors = append(errors, fmt.Errorf("proxy %s: %v", res.proxy.URL, res.err))
	}

	return "", fmt.Errorf("all %d proxy racers failed: %v", len(candidates), errors[len(errors)-1])
}
