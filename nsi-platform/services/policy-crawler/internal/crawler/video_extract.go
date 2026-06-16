package crawler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var ytDlpCookiesFile = os.Getenv("YT_DLP_COOKIES_FILE")

// ytDlpCookiePath is the path to a dynamically generated cookie file (refreshed periodically)
var ytDlpCookiePath string
var ytDlpCookieMu sync.Mutex

func SetYtDlpCookiePath(path string) {
	ytDlpCookieMu.Lock()
	defer ytDlpCookieMu.Unlock()
	ytDlpCookiePath = path
}

func ytDlpCmd(args ...string) *exec.Cmd {
	// Check for dynamically set cookie file (takes precedence over env var)
	ytDlpCookieMu.Lock()
	cookiePath := ytDlpCookiePath
	ytDlpCookieMu.Unlock()

	if cookiePath != "" {
		if _, err := os.Stat(cookiePath); err == nil {
			args = append(args, "--cookies", cookiePath)
		}
	} else if ytDlpCookiesFile != "" {
		if _, err := os.Stat(ytDlpCookiesFile); err == nil {
			args = append(args, "--cookies", ytDlpCookiesFile)
		}
	}
	cmd := exec.Command("yt-dlp", args...)
	return cmd
}

type VideoExtractTask struct {
	RawTextID  int64
	SourceID   string
	VideoURL   string
	Title      string
	RetryCount int
}

type VideoExtractWorker struct {
	store           *DBStore
	filter          *RelevanceFilter
	asr             ASRProvider
	queue           chan VideoExtractTask
	workers         int
	stopCh          <-chan struct{}
	tmpDir          string
	cdpExtractor     *CDPVideoExtractor
	cdpExtractorOnce sync.Once
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

	isManual := strings.HasPrefix(task.SourceID, "DOUYIN-")
	if !isManual && w.filter != nil {
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
	cmd := ytDlpCmd("--write-sub", "--sub-lang", "zh,zh-Hans,zh-CN", "--skip-download",
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

func (w *VideoExtractWorker) getCDPExtractor() *CDPVideoExtractor {
	w.cdpExtractorOnce.Do(func() {
		chromeBin := findChromeBinary()
		if chromeBin != "" {
			w.cdpExtractor = NewCDPVideoExtractor()
		}
	})
	return w.cdpExtractor
}

func (w *VideoExtractWorker) extractViaASR(videoURL, tmpBase string) (string, error) {
	if w.asr == nil {
		return "", fmt.Errorf("ASR provider not configured")
	}

	if cdp := w.getCDPExtractor(); cdp != nil {
		cdpURL, cdpErr := cdp.ExtractVideoURL(videoURL)
		if cdpErr == nil && cdpURL != "" {
			log.Printf("[video-extract] CDP extracted video URL for %s -> %s", videoURL, truncateForLog(cdpURL, 120))
			audioPath := tmpBase + "_cdp.mp3"
			dlReq, dlErr := http.NewRequest("GET", cdpURL, nil)
			if dlErr == nil {
				dlReq.Header.Set("Referer", "https://www.douyin.com/")
				dlReq.Header.Set("User-Agent", "Mozilla/5.0")
				dlResp, dlRespErr := http.DefaultClient.Do(dlReq)
				if dlRespErr == nil && dlResp.StatusCode == 200 {
					f, fErr := os.Create(audioPath)
					if fErr == nil {
						_, copyErr := io.Copy(f, dlResp.Body)
						f.Close()
						dlResp.Body.Close()
						if copyErr == nil {
							return w.asr.Transcribe(audioPath, "")
						}
					} else {
						dlResp.Body.Close()
					}
				} else if dlRespErr == nil {
					dlResp.Body.Close()
				}
			}
			log.Printf("[video-extract] CDP URL download failed, trying ffmpeg")
			ffCmd := exec.Command("ffmpeg",
				"-user_agent", "Mozilla/5.0",
				"-headers", "Referer: https://www.douyin.com/",
				"-i", cdpURL,
				"-vn", "-acodec", "libmp3lame",
				"-ab", "64k", "-ar", "16000", "-ac", "1",
				"-timeout", "30000000",
				"-y", audioPath)
			if ffOut, ffErr := ffCmd.CombinedOutput(); ffErr != nil {
				log.Printf("[video-extract] ffmpeg stderr for %s:\n%s", videoURL, truncateBytes(ffOut, 500))
			} else if _, statErr := os.Stat(audioPath); statErr == nil {
				return w.asr.Transcribe(audioPath, "")
			}
		}
		log.Printf("[video-extract] CDP extraction failed for %s: %v, falling back to yt-dlp", videoURL, cdpErr)
	}

	directAudioURL, err := GetDirectAudioURLFromVideo(videoURL)
	if err != nil {
		log.Printf("[video-extract] failed to get direct audio URL for %s: %v, downloading locally", videoURL, err)
		cmd := ytDlpCmd("-x", "--audio-format", "mp3", "--audio-quality", "5",
			"--output", tmpBase, "--limit-rate", "1M", videoURL)
		out, dlErr := cmd.CombinedOutput()
		if dlErr != nil {
			return "", fmt.Errorf("yt-dlp download: %w: %s", dlErr, string(out))
		}
		audioPath := tmpBase + ".mp3"
		if _, statErr := os.Stat(audioPath); statErr != nil {
			return "", fmt.Errorf("audio file not found at %s", audioPath)
		}
		return w.asr.Transcribe(audioPath, "")
	}

	log.Printf("[video-extract] got direct audio URL, downloading locally for base64 ASR")
	rawAudioPath := tmpBase + "_raw.mp4"
	dlReq, dlReqErr := http.NewRequest("GET", directAudioURL, nil)
	if dlReqErr != nil {
		return "", fmt.Errorf("create download request: %w", dlReqErr)
	}
	dlReq.Header.Set("Referer", "https://www.bilibili.com/")
	dlReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	dlResp, dlRespErr := http.DefaultClient.Do(dlReq)
	if dlRespErr != nil {
		return "", fmt.Errorf("download audio: %w", dlRespErr)
	}
	defer dlResp.Body.Close()
	if dlResp.StatusCode != 200 {
		return "", fmt.Errorf("download audio: HTTP %d", dlResp.StatusCode)
	}
	f, fErr := os.Create(rawAudioPath)
	if fErr != nil {
		return "", fmt.Errorf("create audio file: %w", fErr)
	}
	if _, copyErr := io.Copy(f, dlResp.Body); copyErr != nil {
		f.Close()
		return "", fmt.Errorf("save audio: %w", copyErr)
	}
	f.Close()

	audioPath := tmpBase + "_direct.mp4"
	ffCmd := exec.Command("ffmpeg", "-i", rawAudioPath,
		"-vn", "-acodec", "aac",
		"-ar", "16000", "-ac", "1",
		"-y", audioPath)
	if ffOut, ffErr := ffCmd.CombinedOutput(); ffErr != nil {
		return "", fmt.Errorf("ffmpeg convert: %w: %s", ffErr, string(ffOut))
	}
	return w.asr.Transcribe(audioPath, "")
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
