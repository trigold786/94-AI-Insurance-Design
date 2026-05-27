package crawler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type RobotsChecker struct {
	mu    sync.RWMutex
	cache map[string]*robotsRule
	client *http.Client
}

type robotsRule struct {
	allowed     map[string]bool
	fetched     time.Time
	disallowAll bool
}

func NewRobotsChecker() *RobotsChecker {
	return &RobotsChecker{
		cache:  make(map[string]*robotsRule),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (rc *RobotsChecker) IsAllowed(targetURL, userAgent string) bool {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return true
	}
	rule := rc.getRule(parsed.Scheme + "://" + parsed.Host)
	if rule == nil {
		return true
	}
	if rule.disallowAll {
		return false
	}
	return rule.isAllowed(parsed.Path, userAgent)
}

func (rc *RobotsChecker) getRule(baseURL string) *robotsRule {
	rc.mu.RLock()
	r, ok := rc.cache[baseURL]
	rc.mu.RUnlock()
	if ok && time.Since(r.fetched) < 6*time.Hour {
		return r
	}
	r = rc.fetchRule(baseURL)
	rc.mu.Lock()
	rc.cache[baseURL] = r
	rc.mu.Unlock()
	return r
}

func (rc *RobotsChecker) fetchRule(baseURL string) *robotsRule {
	resp, err := rc.client.Get(baseURL + "/robots.txt")
	if err != nil {
		return &robotsRule{allowed: map[string]bool{}, fetched: time.Now()}
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return &robotsRule{allowed: map[string]bool{}, fetched: time.Now()}
	}
	body, _ := io.ReadAll(resp.Body)
	return parseRobotsTxt(string(body))
}

func parseRobotsTxt(content string) *robotsRule {
	rule := &robotsRule{
		allowed: map[string]bool{},
		fetched: time.Now(),
	}
	var currentUA string
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.SplitN(line, "#", 2)[0]
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "user-agent:") {
			currentUA = strings.TrimSpace(line[len("user-agent:"):])
			continue
		}
		if currentUA != "*" && !strings.EqualFold(currentUA, "Mozilla") {
			continue
		}
		if strings.HasPrefix(lower, "disallow:") {
			path := strings.TrimSpace(line[len("disallow:"):])
			if path == "/" {
				rule.disallowAll = true
				return rule
			}
			if path != "" {
				rule.allowed[path] = false
			}
		}
		if strings.HasPrefix(lower, "allow:") {
			path := strings.TrimSpace(line[len("allow:"):])
			if path != "" {
				rule.allowed[path] = true
			}
		}
	}
	return rule
}

func (r *robotsRule) isAllowed(path, userAgent string) bool {
	longestMatch := ""
	allowed := true
	for pattern, isAllowed := range r.allowed {
		if strings.HasPrefix(path, pattern) && len(pattern) > len(longestMatch) {
			longestMatch = pattern
			allowed = isAllowed
		}
	}
	return allowed
}

func CheckRobotsBeforeCrawl(checker *RobotsChecker, targetURL, userAgent string) error {
	if checker == nil {
		return nil
	}
	if !checker.IsAllowed(targetURL, userAgent) {
		return fmt.Errorf("robots.txt disallows crawling: %s", targetURL)
	}
	return nil
}
