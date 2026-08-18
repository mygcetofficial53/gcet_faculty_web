package service

import (
	"context"
	"io"
	"math/rand"
	"net/http"
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

	// Fetch periodically every 10 minutes
	go GlobalProxyPool.startRotation()
}

// StopProxyPool stops the background fetcher
func StopProxyPool() {
	if GlobalProxyPool != nil && GlobalProxyPool.cancel != nil {
		GlobalProxyPool.cancel()
	}
}

func (p *ProxyPool) startRotation() {
	ticker := time.NewTicker(10 * time.Minute)
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
	var validProxies []string

	for _, line := range lines {
		proxy := strings.TrimSpace(line)
		if proxy == "" {
			continue
		}
		
		// Ensure it has a scheme. If not, assume http.
		if !strings.Contains(proxy, "://") {
			proxy = "http://" + proxy
		}

		proxyLower := strings.ToLower(proxy)
		// Go's net/http transport supports http, https, and socks5
		// We explicitly ignore socks4 since it will cause "unsupported protocol scheme"
		if strings.HasPrefix(proxyLower, "http://") || 
		   strings.HasPrefix(proxyLower, "https://") || 
		   strings.HasPrefix(proxyLower, "socks5://") {
			validProxies = append(validProxies, proxy)
		}
	}

	if len(validProxies) > 0 {
		p.mu.Lock()
		p.proxies = validProxies
		p.mu.Unlock()
		logger.Log.Infof("ProxyPool: loaded %d valid proxies", len(validProxies))
	} else {
		logger.Log.Warn("ProxyPool: fetched list contained 0 valid proxies")
	}
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
