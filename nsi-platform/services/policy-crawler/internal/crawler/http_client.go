package crawler

import (
	"crypto/tls"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HTTPClientConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	Timeout    time.Duration
	ProxyURL   string
	UserAgent  string
}

func DefaultHTTPClientConfig() HTTPClientConfig {
	return HTTPClientConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		Timeout:    30 * time.Second,
		UserAgent:  govUserAgents[0],
	}
}

type ResilientHTTPClient struct {
	config HTTPClientConfig
	client *http.Client
}

func NewResilientHTTPClient(cfg HTTPClientConfig) *ResilientHTTPClient {
	tr := &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableKeepAlives:   false,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	if cfg.ProxyURL != "" {
		if proxyURL, err := url.Parse(cfg.ProxyURL); err == nil {
			tr.Proxy = http.ProxyURL(proxyURL)
		}
	}
	return &ResilientHTTPClient{
		config: cfg,
		client: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: tr,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
	}
}

func (c *ResilientHTTPClient) Get(urlStr string) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * c.config.BaseDelay
			time.Sleep(delay)
		}
		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("User-Agent", c.config.UserAgent)

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			if !isRetryableError(err) {
				return nil, err
			}
			continue
		}
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("HTTP %d from %s", resp.StatusCode, urlStr)
			resp.Body.Close()
			continue
		}
		return resp, nil
	}
	return nil, fmt.Errorf("after %d retries: %w", c.config.MaxRetries, lastErr)
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline") {
		return true
	}
	if strings.Contains(msg, "connection refused") || strings.Contains(msg, "EOF") {
		return true
	}
	if strings.Contains(msg, "temporary") || strings.Contains(msg, "TLS handshake") {
		return true
	}
	if strings.Contains(msg, "no such host") {
		return false
	}
	return true
}

func (c *ResilientHTTPClient) HTTPClient() *http.Client {
	return c.client
}
