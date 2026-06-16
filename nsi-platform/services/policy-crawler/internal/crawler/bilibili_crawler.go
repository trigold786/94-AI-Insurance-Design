package crawler

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type BilibiliCrawler struct {
	config    SourceConfig
	filter    *RelevanceFilter
	maxItems  int
	processed map[string]bool
	client    *http.Client
}

func NewBilibiliCrawler(cfg SourceConfig, filter *RelevanceFilter) *BilibiliCrawler {
	maxItems := 50
	if strings.HasPrefix(cfg.SourceID, "BILIBILI-") {
		maxItems = 9999
	}
	return &BilibiliCrawler{
		config:    cfg,
		filter:    filter,
		maxItems:  maxItems,
		processed: make(map[string]bool),
		client:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (b *BilibiliCrawler) SourceID() string   { return b.config.SourceID }
func (b *BilibiliCrawler) SourceLevel() string { return b.config.SourceLevel }

func (b *BilibiliCrawler) Interval() time.Duration {
	if b.config.IntervalSec <= 0 {
		return 168 * time.Hour
	}
	return time.Duration(b.config.IntervalSec) * time.Second
}

func (b *BilibiliCrawler) Fetch() ([]*CrawlResult, error) {
	if b.config.SourceURL == "" {
		return nil, nil
	}

	urls := parseBilibiliURLs(b.config.SourceURL)
	var videoURLs []string
	for _, u := range urls {
		if strings.HasPrefix(u, "search:") {
			keyword := strings.TrimPrefix(u, "search:")
			keyword = strings.TrimSpace(keyword)
			if keyword == "" {
				continue
			}
			discovered, err := b.searchVideos(keyword)
			if err != nil {
				log.Printf("[bilibili] search %q error: %v", keyword, err)
				continue
			}
			videoURLs = append(videoURLs, discovered...)
		} else if isBilibiliSpaceURL(u) {
			uid := extractBilibiliUID(u)
			if uid == "" {
				log.Printf("[bilibili] cannot extract UID from %s", u)
				continue
			}
			discovered, err := b.discoverVideosFromSpace(uid)
			if err != nil {
				log.Printf("[bilibili] discover space %s error: %v", u, err)
				continue
			}
			videoURLs = append(videoURLs, discovered...)
		} else if isBilibiliVideoURL(u) {
			videoURLs = append(videoURLs, u)
		} else {
			log.Printf("[bilibili] unrecognized URL format: %s", u)
		}
	}

	var results []*CrawlResult
	for _, u := range videoURLs {
		if b.processed[u] {
			continue
		}
		b.processed[u] = true

		result, err := b.fetchVideoDetail(u)
		if err != nil {
			log.Printf("[bilibili] fetch %s error: %v", u, err)
			continue
		}
		if result != nil {
			results = append(results, result)
		}
		if len(results) >= b.maxItems {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	return results, nil
}

func (b *BilibiliCrawler) bilibiliAPIGet(apiURL string, target interface{}) error {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.bilibili.com/")

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	var apiResp struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("parse API response: %w (body=%s)", err, truncateBytes(body, 200))
	}
	if apiResp.Code != 0 {
		return fmt.Errorf("API error code=%d msg=%s", apiResp.Code, apiResp.Message)
	}
	return json.Unmarshal(apiResp.Data, target)
}

func (b *BilibiliCrawler) discoverVideosFromSpace(uid string) ([]string, error) {
	log.Printf("[bilibili] discovering videos from space uid=%s", uid)
	var allURLs []string
	for page := 1; page <= 10; page++ {
		apiURL := fmt.Sprintf("https://api.bilibili.com/x/space/arc/search?mid=%s&ps=50&pn=%d", uid, page)
		var data struct {
			List struct {
				Vlist []struct {
					BVID string `json:"bvid"`
				} `json:"vlist"`
			} `json:"list"`
		}
		if err := b.bilibiliAPIGet(apiURL, &data); err != nil {
			if page == 1 {
				return nil, fmt.Errorf("space API page %d: %w", page, err)
			}
			break
		}
		if len(data.List.Vlist) == 0 {
			break
		}
		for _, v := range data.List.Vlist {
			allURLs = append(allURLs, "https://www.bilibili.com/video/"+v.BVID)
		}
		log.Printf("[bilibili] space uid=%s page=%d found %d videos (total=%d)", uid, page, len(data.List.Vlist), len(allURLs))
		time.Sleep(300 * time.Millisecond)
	}
	if len(allURLs) == 0 {
		return nil, fmt.Errorf("no videos found for space uid=%s", uid)
	}
	log.Printf("[bilibili] space uid=%s total discovered %d videos", uid, len(allURLs))
	return allURLs, nil
}

func (b *BilibiliCrawler) searchVideos(keyword string) ([]string, error) {
	log.Printf("[bilibili] searching videos for keyword=%q", keyword)
	apiURL := fmt.Sprintf("https://api.bilibili.com/x/web-interface/search/type?search_type=video&keyword=%s&page=1&page_size=50", url.QueryEscape(keyword))

	var data struct {
		Result []struct {
			Type string `json:"type"`
			BVID string `json:"bvid"`
		} `json:"result"`
	}
	if err := b.bilibiliAPIGet(apiURL, &data); err != nil {
		return nil, fmt.Errorf("search API: %w", err)
	}

	var urls []string
	for _, v := range data.Result {
		if v.Type != "video" || v.BVID == "" {
			continue
		}
		urls = append(urls, "https://www.bilibili.com/video/"+v.BVID)
	}
	log.Printf("[bilibili] search %q found %d videos", keyword, len(urls))
	return urls, nil
}

func (b *BilibiliCrawler) fetchVideoDetail(videoURL string) (*CrawlResult, error) {
	bvid := extractBilibiliBVID(videoURL)
	if bvid == "" {
		return nil, fmt.Errorf("cannot extract BVID from %s", videoURL)
	}

	apiURL := fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?bvid=%s", bvid)
	var data struct {
		BVID  string `json:"bvid"`
		Title string `json:"title"`
		Desc  string `json:"desc"`
		Owner struct {
			Name string `json:"name"`
			MID  int    `json:"mid"`
		} `json:"owner"`
	}
	if err := b.bilibiliAPIGet(apiURL, &data); err != nil {
		return nil, fmt.Errorf("video detail API: %w", err)
	}

	title := data.Title
	desc := data.Desc
	content := title
	if desc != "" {
		content = title + "\n" + desc
	}
	if content == "" {
		return nil, nil
	}

	searchText := title + " " + desc
	isManual := strings.HasPrefix(b.config.SourceID, "BILIBILI-")
	if !isManual && b.filter != nil {
		score, matched := b.filter.Score(searchText, b.config.SourceID, "bilibili")
		threshold := b.filter.MinScore(b.config.SourceID, "level1")
		if score < threshold {
			log.Printf("[bilibili] filtered out %s: relevance score %d < threshold %d (matched: %v)", videoURL, score, threshold, matched)
			return nil, nil
		}
		log.Printf("[bilibili] passed relevance filter for %s: score=%d matched=%v", videoURL, score, matched)
	}

	hash := sha256.Sum256([]byte(videoURL))
	return &CrawlResult{
		SourceID:          b.config.SourceID,
		SourceLevel:       b.config.SourceLevel,
		RawText:           content,
		Title:             title,
		SourceURL:         videoURL,
		FetchedAt:         time.Now(),
		VersionHash:       fmt.Sprintf("%x", hash),
		VideoURL:          videoURL,
		NeedsVideoExtract: true,
		ContentType:       "video-meta",
	}, nil
}

func parseBilibiliURLs(rawURL string) []string {
	var urls []string
	seen := make(map[string]bool)
	for _, part := range strings.Split(rawURL, "\n") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, "search:") {
			kw := strings.TrimPrefix(part, "search:")
			kw = strings.TrimSpace(kw)
			if kw != "" && !seen[kw] {
				seen[kw] = true
				urls = append(urls, "search:"+kw)
			}
		} else if strings.HasPrefix(part, "http") && (strings.Contains(part, "bilibili.com") || strings.Contains(part, "b23.tv")) {
			if !seen[part] {
				seen[part] = true
				urls = append(urls, part)
			}
		}
	}
	if len(urls) == 0 {
		raw := strings.TrimSpace(rawURL)
		if raw != "" && !seen[raw] && (strings.Contains(raw, "bilibili.com") || strings.Contains(raw, "b23.tv")) {
			urls = append(urls, raw)
		}
	}
	return urls
}

var bvidRe = regexp.MustCompile(`BV[a-zA-Z0-9]{10,12}`)

func extractBilibiliBVID(url string) string {
	return bvidRe.FindString(url)
}

func isBilibiliSpaceURL(u string) bool {
	return strings.Contains(u, "space.bilibili.com")
}

func isBilibiliVideoURL(u string) bool {
	return strings.Contains(u, "bilibili.com/video/") || strings.Contains(u, "b23.tv")
}

func extractBilibiliUID(spaceURL string) string {
	re := regexp.MustCompile(`space\.bilibili\.com/(\d+)`)
	matches := re.FindStringSubmatch(spaceURL)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
