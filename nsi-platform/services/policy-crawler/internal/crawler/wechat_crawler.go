package crawler

import (
	"crypto/sha256"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

var wechatArticleRe = regexp.MustCompile(`https?://mp\.weixin\.qq\.com/s[^\s"'<>]+`)
var sogouResultRe = regexp.MustCompile(`https?://mp\.weixin\.qq\.com/[^\s"'<>]+`)

type WeChatCrawler struct {
	config    SourceConfig
	renderer  PageRenderer
	filter    *RelevanceFilter
	maxItems  int
	processed map[string]bool
}

func NewWeChatCrawler(cfg SourceConfig, filter *RelevanceFilter) *WeChatCrawler {
	return &WeChatCrawler{
		config:    cfg,
		filter:    filter,
		maxItems:  20,
		processed: make(map[string]bool),
	}
}

func (w *WeChatCrawler) SetRenderer(r PageRenderer) { w.renderer = r }
func (w *WeChatCrawler) SourceID() string            { return w.config.SourceID }
func (w *WeChatCrawler) SourceLevel() string          { return w.config.SourceLevel }

func (w *WeChatCrawler) Interval() time.Duration {
	if w.config.IntervalSec <= 0 {
		return 168 * time.Hour
	}
	return time.Duration(w.config.IntervalSec) * time.Second
}

func (w *WeChatCrawler) Fetch() ([]*CrawlResult, error) {
	if w.config.SourceURL == "" {
		return nil, nil
	}
	if w.renderer == nil {
		log.Printf("[wechat] %s: Chrome renderer not available, skipping", w.config.SourceID)
		return nil, nil
	}

	urls := parseWeChatURLs(w.config.SourceURL)

	var articleURLs []string
	var keywords []string
	for _, u := range urls {
		if isWeChatArticleURL(u) {
			articleURLs = append(articleURLs, u)
		} else {
			keywords = append(keywords, u)
		}
	}

	for _, kw := range keywords {
		discovered := w.discoverArticles(kw)
		articleURLs = append(articleURLs, discovered...)
	}

	if len(articleURLs) > 0 {
		bizID := extractBizID(articleURLs[0])
		if bizID != "" {
			bizArticles := w.discoverByBiz(bizID)
			articleURLs = append(articleURLs, bizArticles...)
		}
	}

	seen := make(map[string]bool)
	var deduped []string
	for _, u := range articleURLs {
		u = cleanWeChatURL(u)
		if !seen[u] {
			seen[u] = true
			deduped = append(deduped, u)
		}
	}

	var results []*CrawlResult
	for _, u := range deduped {
		if w.processed[u] {
			continue
		}
		w.processed[u] = true

		result, err := w.fetchArticle(u)
		if err != nil {
			log.Printf("[wechat] fetch %s error: %v", u, err)
			continue
		}
		if result == nil {
			continue
		}

		if w.filter != nil {
			score, matched := w.filter.Score(result.Title+" "+result.RawText, w.config.SourceID, "wechat")
			threshold := w.filter.MinScore(w.config.SourceID, "level1")
			if score < threshold {
				log.Printf("[wechat] filtered out %s: score %d < threshold %d (matched: %v)", u, score, threshold, matched)
				continue
			}
		}

		results = append(results, result)
		if len(results) >= w.maxItems {
			break
		}
		time.Sleep(3 * time.Second)
	}
	return results, nil
}

func (w *WeChatCrawler) discoverArticles(keyword string) []string {
	urls := w.discoverArticlesBaidu(keyword)
	if len(urls) > 0 {
		return urls
	}
	urls = w.discoverArticlesSogou(keyword)
	if len(urls) > 0 {
		return urls
	}
	urls = w.discoverArticlesBing(keyword)
	return urls
}

func (w *WeChatCrawler) discoverArticlesBing(keyword string) []string {
	searchURL := fmt.Sprintf("https://www.bing.com/search?q=site%%3Amp.weixin.qq.com%%20%s", keyword)
	log.Printf("[wechat] discovering articles via Bing: %s", keyword)

	html, err := w.renderer.RenderWithVirtualTime(searchURL, 15000)
	if err != nil {
		log.Printf("[wechat] Bing search render error: %v", err)
		return nil
	}

	if strings.Contains(html, "验证") && strings.Contains(html, "captcha") {
		log.Printf("[wechat] Bing search CAPTCHA detected, skipping")
		return nil
	}

	return w.extractWeChatURLs(html, "Bing", keyword)
}

func (w *WeChatCrawler) discoverByBiz(bizID string) []string {
	searchURL := fmt.Sprintf("https://www.bing.com/search?q=site%%3Amp.weixin.qq.com%%20__biz%%3D%s", bizID)
	log.Printf("[wechat] discovering articles by biz: %s", bizID)
	html, err := w.renderer.RenderWithVirtualTime(searchURL, 20000)
	if err != nil {
		log.Printf("[wechat] biz search render error: %v", err)
		return nil
	}
	return w.extractWeChatURLs(html, "BizSearch", bizID)
}

func extractBizID(articleURL string) string {
	idx := strings.Index(articleURL, "__biz=")
	if idx == -1 {
		idx = strings.Index(articleURL, "__biz%3D")
	}
	if idx == -1 {
		return ""
	}
	start := idx
	if strings.Contains(articleURL[start:start+6], "%3D") {
		start = idx + 8
	} else {
		start = idx + 6
	}
	end := strings.IndexAny(articleURL[start:], "&#")
	if end == -1 {
		end = len(articleURL)
	} else {
		end += start
	}
	return articleURL[start:end]
}

func (w *WeChatCrawler) discoverArticlesBaidu(keyword string) []string {
	searchURL := fmt.Sprintf("https://www.baidu.com/s?wd=site%%3Amp.weixin.qq.com%%20%s&rn=20", keyword)
	log.Printf("[wechat] discovering articles via Baidu: %s", keyword)

	html, err := w.renderer.RenderWithVirtualTime(searchURL, 15000)
	if err != nil {
		log.Printf("[wechat] Baidu search render error: %v", err)
		return nil
	}

	if strings.Contains(html, "验证码") || strings.Contains(html, "antispider") {
		log.Printf("[wechat] Baidu CAPTCHA detected, skipping")
		return nil
	}

	return w.extractWeChatURLs(html, "Baidu", keyword)
}

func (w *WeChatCrawler) extractWeChatURLs(html, engine, keyword string) []string {
	matches := wechatArticleRe.FindAllString(html, -1)
	seen := make(map[string]bool)
	var urls []string
	for _, u := range matches {
		u = cleanWeChatURL(u)
		if !seen[u] {
			seen[u] = true
			urls = append(urls, u)
		}
	}
	log.Printf("[wechat] %s discovered %d articles for %q", engine, len(urls), keyword)
	return urls
}

func (w *WeChatCrawler) discoverArticlesSogou(keyword string) []string {
	searchURL := fmt.Sprintf("https://weixin.sogou.com/weixin?type=2&query=%s", keyword)
	log.Printf("[wechat] fallback: discovering articles via Sogou: %s", keyword)

	html, err := w.renderer.RenderWithVirtualTime(searchURL, 15000)
	if err != nil {
		log.Printf("[wechat] Sogou search render error: %v", err)
		return nil
	}

	if strings.Contains(html, "用户验证") || strings.Contains(html, "antispider") {
		log.Printf("[wechat] Sogou CAPTCHA detected, skipping")
		return nil
	}

	return w.extractWeChatURLs(html, "Sogou", keyword)
}

func (w *WeChatCrawler) fetchArticle(articleURL string) (*CrawlResult, error) {
	html, err := w.renderer.RenderWithVirtualTime(articleURL, 25000)
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}

	if strings.Contains(html, "环境异常") || (strings.Contains(html, "验证") && !strings.Contains(html, "js_content")) {
		return nil, fmt.Errorf("anti-bot page detected")
	}

	title := extractWeChatTitle(html)
	content := extractWeChatContent(html)
	if content == "" {
		content = extractWeChatTextFallback(html)
	}
	if content == "" {
		return nil, nil
	}

	hash := sha256.Sum256([]byte(articleURL))
	return &CrawlResult{
		SourceID:    w.config.SourceID,
		SourceLevel: w.config.SourceLevel,
		RawText:     content,
		Title:       title,
		SourceURL:   articleURL,
		FetchedAt:   time.Now(),
		VersionHash: fmt.Sprintf("%x", hash),
	}, nil
}

func cleanWeChatURL(u string) string {
	u = strings.ReplaceAll(u, "&amp;", "&")
	if idx := strings.Index(u, "&chksm="); idx != -1 {
		u = u[:idx]
	}
	return strings.TrimSpace(u)
}

func isWeChatArticleURL(u string) bool {
	return strings.Contains(u, "mp.weixin.qq.com")
}

func parseWeChatURLs(rawURL string) []string {
	var urls []string
	seen := make(map[string]bool)
	for _, part := range strings.Split(rawURL, "\n") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, "http") && isWeChatArticleURL(part) {
			if !seen[part] {
				seen[part] = true
				urls = append(urls, part)
			}
		} else if strings.HasPrefix(part, "keyword:") {
			kw := strings.TrimPrefix(part, "keyword:")
			kw = strings.TrimSpace(kw)
			if kw != "" && !seen[kw] {
				seen[kw] = true
				urls = append(urls, kw)
			}
		} else if !strings.HasPrefix(part, "http") {
			if !seen[part] {
				seen[part] = true
				urls = append(urls, part)
			}
		}
	}
	return urls
}

func extractWeChatTitle(html string) string {
	markers := []string{
		`<h1 class="rich_media_title" id="activity-name">`,
		`<h1 class="rich_media_title">`,
		`property="og:title" content="`,
		`<title>`,
	}
	for _, m := range markers {
		idx := strings.Index(html, m)
		if idx == -1 {
			continue
		}
		start := idx + len(m)
		end := strings.Index(html[start:], "<")
		if end == -1 {
			end = strings.Index(html[start:], `"`)
		}
		if end == -1 {
			continue
		}
		title := strings.TrimSpace(html[start : start+end])
		title = strings.ReplaceAll(title, "-微信公众平台", "")
		title = strings.ReplaceAll(title, "_微信公众号", "")
		if title != "" {
			return title
		}
	}
	return ""
}

func extractWeChatContent(html string) string {
	contentStart := strings.Index(html, `id="js_content"`)
	if contentStart == -1 {
		contentStart = strings.Index(html, `class="rich_media_content"`)
	}
	if contentStart == -1 {
		return ""
	}

	divStart := strings.LastIndex(html[:contentStart], "<")
	if divStart == -1 {
		divStart = contentStart
	}

	depth := 0
	i := divStart
	for i < len(html) {
		if html[i] == '<' {
			if i+1 < len(html) && html[i+1] == '/' {
				depth--
				if depth <= 0 {
					end := strings.Index(html[i:], ">")
					if end != -1 {
						content := html[divStart : i+end+1]
						return stripHTMLTags(content)
					}
				}
			} else if !strings.HasPrefix(html[i:], "<br") && !strings.HasPrefix(html[i:], "<img") {
				depth++
			}
		}
		i++
	}

	content := html[divStart:]
	if len(content) > 50000 {
		content = content[:50000]
	}
	return stripHTMLTags(content)
}

func extractWeChatTextFallback(html string) string {
	stripTags := regexp.MustCompile(`<[^>]*>`)
	text := stripTags.ReplaceAllString(html, " ")
	text = strings.Join(strings.Fields(text), " ")
	if len(text) > 500 {
		text = text[:500]
	}
	return strings.TrimSpace(text)
}

func stripHTMLTags(html string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(html, "\n")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", `"`)
	lines := strings.Split(text, "\n")
	var cleaned []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}
	return strings.Join(cleaned, "\n")
}
