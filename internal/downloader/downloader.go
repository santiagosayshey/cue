package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Downloader interface {
	Download(url string, destPath string, filename string) error
}

type Youtube struct {
	mu           sync.Mutex
	lastDownload time.Time
}

type GDrive struct{}
type HTTP struct{}

func (yt *Youtube) Download(url string, destPath string, filename string) error {
	yt.mu.Lock()
	defer yt.mu.Unlock()

	if !yt.lastDownload.IsZero() {
		wait := 5*time.Second - time.Since(yt.lastDownload)
		if wait > 0 {
			time.Sleep(wait)
		}
	}

	cmd := exec.Command("yt-dlp",
		"--quiet",
		"--no-warnings",
		"--no-progress",
		"--js-runtimes", "node",
		"-x", "--audio-format", "mp3",
		"-o", filename, url)
	cmd.Dir = destPath
	out, err := cmd.CombinedOutput()
	yt.lastDownload = time.Now()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return fmt.Errorf("yt-dlp: %w", err)
		}
		lines := strings.Split(msg, "\n")
		return fmt.Errorf("yt-dlp: %w: %s", err, strings.TrimSpace(lines[len(lines)-1]))
	}
	return nil
}

func (gd GDrive) Download(url string, destPath string, filename string) error {
	resp, err := http.Get("https://drive.google.com/uc?export=download&id=" + strings.Split(url, "/")[5])
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	file, err := os.Create(filepath.Join(destPath, filename))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	return err
}

func (h HTTP) Download(url string, destPath string, filename string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "cue/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %s: %s", resp.Status, url)
	}
	file, err := os.Create(filepath.Join(destPath, filename))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	return err
}

func NewDownloaders() map[string]Downloader {
	return map[string]Downloader{
		"youtube": &Youtube{},
		"gdrive":  GDrive{},
		"http":    HTTP{},
	}
}
