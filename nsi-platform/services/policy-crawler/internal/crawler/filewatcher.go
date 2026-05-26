package crawler

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileWatcherCrawler 本地政策文件监控爬虫（离线/演示模式）
// 监控指定目录中的 .txt 文件，读取内容作为政策原文
type FileWatcherCrawler struct {
	config    SourceConfig
	watchDir  string
	processed map[string]bool
}

func NewFileWatcherCrawler(cfg SourceConfig, watchDir string) *FileWatcherCrawler {
	return &FileWatcherCrawler{
		config:    cfg,
		watchDir:  watchDir,
		processed: make(map[string]bool),
	}
}

func (f *FileWatcherCrawler) SourceID() string  { return f.config.SourceID }
func (f *FileWatcherCrawler) SourceLevel() string { return f.config.SourceLevel }

func (f *FileWatcherCrawler) Interval() time.Duration {
	if f.config.IntervalSec <= 0 {
		return 60 * time.Second // 文件模式默认每分钟检查
	}
	return time.Duration(f.config.IntervalSec) * time.Second
}

func (f *FileWatcherCrawler) Fetch() ([]*CrawlResult, error) {
	if err := os.MkdirAll(f.watchDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create watch dir: %w", err)
	}

	entries, err := os.ReadDir(f.watchDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read watch dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".txt" {
			continue
		}
		if f.processed[entry.Name()] {
			continue
		}

		data, err := os.ReadFile(filepath.Join(f.watchDir, entry.Name()))
		if err != nil {
			continue
		}

		f.processed[entry.Name()] = true
		hash := sha256.Sum256(data)

		return []*CrawlResult{{
			SourceID:    f.config.SourceID,
			SourceLevel: f.config.SourceLevel,
			RawText:     string(data),
			Title:       entry.Name(),
			SourceURL:   "file://" + filepath.Join(f.watchDir, entry.Name()),
			FetchedAt:   time.Now(),
			VersionHash: fmt.Sprintf("%x", hash),
		}}, nil
	}

	return nil, nil // 没有新文件
}
