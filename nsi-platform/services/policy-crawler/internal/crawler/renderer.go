package crawler

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

var chromeMu sync.Mutex

var (
	chromeBinOnce sync.Once
	chromeBinPath string
)

type PageRenderer interface {
	Render(url string) (string, error)
	RenderWithVirtualTime(url string, budgetMs int) (string, error)
}

var chromeCandidates = []string{
	"chromium-browser",
	"google-chrome",
	"google-chrome-stable",
	"google-chrome-unstable",
	"chromium",
	"chrome",
	"chrome.exe",
}

func findChromeBinary() string {
	chromeBinOnce.Do(func() {
		if p := os.Getenv("CHROME_PATH"); p != "" {
			if _, err := os.Stat(p); err == nil {
				chromeBinPath = p
				return
			}
			log.Printf("[chrome] CHROME_PATH=%s not found, falling back to LookPath", p)
		}
		for _, name := range chromeCandidates {
			if p, err := exec.LookPath(name); err == nil {
				chromeBinPath = p
				return
			}
		}
	})
	return chromeBinPath
}

type ChromeRenderer struct {
	timeout time.Duration
}

func NewChromeRenderer() *ChromeRenderer {
	return &ChromeRenderer{timeout: 60 * time.Second}
}

func (r *ChromeRenderer) Render(url string) (string, error) {
	return r.RenderWithVirtualTime(url, 0)
}

func (r *ChromeRenderer) RenderWithVirtualTime(url string, budgetMs int) (string, error) {
	chromeMu.Lock()
	defer chromeMu.Unlock()

	chromeBin := findChromeBinary()
	if chromeBin == "" {
		return "", fmt.Errorf("chrome binary not found: tried %v and CHROME_PATH", chromeCandidates)
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	args := []string{
		"--headless",
		"--disable-gpu",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--dump-dom",
	}
	if budgetMs > 0 {
		args = append(args, fmt.Sprintf("--virtual-time-budget=%d", budgetMs))
	}
	args = append(args, url)

	cmd := exec.CommandContext(ctx, chromeBin, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("chrome render: %w (stderr: %s)", err, trimStr(stderr.String(), 200))
	}

	out := stdout.Bytes()
	if len(out) < 100 {
		return "", fmt.Errorf("rendered content too short (%d bytes)", len(out))
	}

	return string(out), nil
}

func trimStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
