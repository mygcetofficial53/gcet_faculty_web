package service

import (
	"context"
	"crypto/tls"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"gcet-web-backend/internal/logger"
)

// Global Proxy Pool
var GlobalProxyPool *ProxyPool

func init() {
	rand.Seed(time.Now().UnixNano())
}

type ProxyPool struct {
	url     string
	proxies []string
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
	var candidateProxies []string

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

		if strings.HasPrefix(proxyLower, "http://") || strings.HasPrefix(proxyLower, "socks5://") {
			candidateProxies = append(candidateProxies, proxy)
		}
	}

	logger.Log.Infof("ProxyPool: Fetched %d candidate proxies. Starting health checks...", len(candidateProxies))
	
	// Test the proxies in the background so we don't block
	go p.testAndSetProxies(candidateProxies)
}

func (p *ProxyPool) testAndSetProxies(candidates []string) {
	var validProxies []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Limit concurrency to 50 parallel checks
	semaphore := make(chan struct{}, 50)

	for _, proxy := range candidates {
		wg.Add(1)
		go func(px string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if checkProxyHealth(px) {
				mu.Lock()
				validProxies = append(validProxies, px)
				mu.Unlock()
			}
		}(proxy)
	}

	wg.Wait()

	if len(validProxies) > 0 {
		p.mu.Lock()
		p.proxies = validProxies
		p.mu.Unlock()
		logger.Log.Infof("ProxyPool: Kept %d healthy, fast proxies from the candidate list", len(validProxies))
	} else {
		logger.Log.Warn("ProxyPool: ALL candidate proxies were dead or too slow! Keeping old list if available.")
	}
}

func checkProxyHealth(proxyStr string) bool {
	proxyURL, err := url.Parse(proxyStr)
	if err != nil {
		return false
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = http.ProxyURL(proxyURL)
	transport.ForceAttemptHTTP2 = false
	transport.TLSNextProto = make(map[string]func(authority string, c *tls.Conn) http.RoundTripper)
	
	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second, // VERY aggressive timeout. Must be fast!
	}

	req, err := http.NewRequest("HEAD", "http://202.129.240.148:8080/GIS", nil)
	if err != nil {
		return false
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200 || resp.StatusCode == 302 // GMS returns 200 or 302
}

// GetRandomProxy returns a random proxy from the pool, or empty string if pool is empty
func (p *ProxyPool) GetRandomProxy() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.proxies) == 0 {
		return ""
	}

	idx := rand.Intn(len(p.proxies))
	return p.proxies[idx]
}
