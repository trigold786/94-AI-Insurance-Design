package crawler

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type VideoExtractTask struct {
	RawTextID  int64
	SourceID   string
	VideoURL   string
	Title      string
	RetryCount int
}

type VideoExtractWorker struct {
	store   *DBStore
	filter  *RelevanceFilter
	asr     ASRProvider
	queue   chan VideoExtractTask
	workers int
	stopCh  <-chan struct{}
	tmpDir  string
}

func NewVideoExtractWorker(store *DBStore, filter *RelevanceFilter, asr ASRProvider, workers int) *VideoExtractWorker {
	if workers <= 0 {
		workers = 2
	}
	tmpDir := "/tmp/video-extract"
	os.MkdirAll(tmpDir, 0755)
	return &VideoExtractWorker{
		store:   store,
		filter:  filter,
		asr:     asr,
		queue:   make(chan VideoExtractTask, 100),
		workers: workers,
		tmpDir:  tmpDir,
	}
}

func (w *VideoExtractWorker) Queue() chan VideoExtractTask { return w.queue }

func (w *VideoExtractWorker) Start() {
	for i := 0; i < w.workers; i++ {
		go w.run(i)
	}
	log.Printf("[video-extract] started %d workers", w.workers)
}

func (w *VideoExtractWorker) run(id int) {
	for {
		select {
		case task := <-w.queue:
			w.process(task)
		case <-w.stopCh:
			return
		}
	}
}

func (w *VideoExtractWorker) process(task VideoExtractTask) {
	log.Printf("[video-extract] processing raw_text=%d url=%s retry=%d", task.RawTextID, task.VideoURL, task.RetryCount)
	w.store.SetVideoExtractStatus(task.RawTextID, "processing")

	transcript, err := w.extractTranscript(task.VideoURL)
	if err != nil {
		log.Printf("[video-extract] transcript error for %s: %v", task.VideoURL, err)
		w.handleFailure(task, err)
		return
	}

	enriched := fmt.Sprintf("【标题】%s\n【视频转录】%s", task.Title, transcript)

	if w.filter != nil {
		score, matched := w.filter.Score(enriched, task.SourceID, "level2")
		threshold := w.filter.MinScore(task.SourceID, "level2")
		if score < threshold {
			log.Printf("[video-extract] discarded raw_text=%d score=%d<threshold=%d matched=%v", task.RawTextID, score, threshold, matched)
			w.store.UpdateRawTextContent(task.RawTextID, enriched)
			w.store.SetVideoExtractStatus(task.RawTextID, "discarded")
			w.store.MarkExtractedByID(task.RawTextID)
			return
		}
		log.Printf("[video-extract] passed L2 filter raw_text=%d score=%d", task.RawTextID, score)
	}

	w.store.UpdateRawTextContent(task.RawTextID, enriched)
	w.store.SetVideoExtractStatus(task.RawTextID, "done")
	log.Printf("[video-extract] done raw_text=%d enriched=%d bytes", task.RawTextID, len(enriched))
}

func (w *VideoExtractWorker) extractTranscript(videoURL string) (string, error) {
	tmpBase := filepath.Join(w.tmpDir, fmt.Sprintf("%d", time.Now().UnixNano()))
	defer w.cleanup(tmpBase)

	subtitle, err := w.extractSubtitle(videoURL, tmpBase)
	if err == nil && subtitle != "" {
		log.Printf("[video-extract] got subtitle for %s (%d chars)", videoURL, len(subtitle))
		return subtitle, nil
	}
	log.Printf("[video-extract] no subtitle for %s, falling back to ASR: %v", videoURL, err)
	return w.extractViaASR(videoURL, tmpBase)
}

func (w *VideoExtractWorker) extractSubtitle(videoURL, tmpBase string) (string, error) {
	cmd := exec.Command("yt-dlp", "--write-sub", "--sub-lang", "zh,zh-Hans,zh-CN", "--skip-download",
		"--convert-subs", "srt", "--output", tmpBase, videoURL)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp subtitle: %w: %s", err, string(out))
	}
	srtPath := tmpBase + ".srt"
	vttPath := tmpBase + ".vtt"
	data, err := os.ReadFile(srtPath)
	if err != nil {
		data, err = os.ReadFile(vttPath)
		if err != nil {
			return "", fmt.Errorf("no subtitle file found")
		}
	}
	text := cleanSubtitle(string(data))
	if len(text) < 20 {
		return "", fmt.Errorf("subtitle too short: %d chars", len(text))
	}
	return text, nil
}

func (w *VideoExtractWorker) extractViaASR(videoURL, tmpBase string) (string, error) {
	cmd := exec.Command("yt-dlp", "-x", "--audio-format", "mp3", "--audio-quality", "5",
		"--output", tmpBase, "--ratelimit", "1M", videoURL)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp download: %w: %s", err, string(out))
	}
	audioPath := tmpBase + ".mp3"
	if _, err := os.Stat(audioPath); err != nil {
		return "", fmt.Errorf("audio file not found at %s", audioPath)
	}
	if w.asr == nil {
		return "", fmt.Errorf("ASR provider not configured")
	}
	return w.asr.Transcribe(audioPath)
}

func (w *VideoExtractWorker) handleFailure(task VideoExtractTask, err error) {
	if task.RetryCount < 3 {
		task.RetryCount++
		log.Printf("[video-extract] retrying raw_text=%d (attempt %d)", task.RawTextID, task.RetryCount)
		time.AfterFunc(time.Duration(task.RetryCount)*10*time.Second, func() {
			w.queue <- task
		})
	} else {
		w.store.SetVideoExtractStatus(task.RawTextID, "failed")
		log.Printf("[video-extract] giving up raw_text=%d after %d retries: %v", task.RawTextID, task.RetryCount, err)
	}
}

func (w *VideoExtractWorker) cleanup(base string) {
	patterns := []string{base + ".*", base}
	for _, p := range patterns {
		matches, _ := filepath.Glob(p)
		for _, f := range matches {
			os.Remove(f)
		}
	}
}

func cleanSubtitle(srt string) string {
	var lines []string
	for _, line := range strings.Split(srt, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, "-->") {
			continue
		}
		if len(line) > 0 && line[0] >= '0' && line[0] <= '9' && !strings.Contains(line, " ") && len(line) <= 5 {
			continue
		}
		line = strings.ReplaceAll(line, "<b>", "")
		line = strings.ReplaceAll(line, "</b>", "")
		line = strings.ReplaceAll(line, "<i>", "")
		line = strings.ReplaceAll(line, "</i>", "")
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, " ")
}
