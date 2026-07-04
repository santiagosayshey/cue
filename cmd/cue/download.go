package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Downloader interface {
	Download(url string, destPath string) error
}

type Youtube struct{}
type GDrive struct{}

func (yt Youtube) Download(url string, destPath string) error {
	cmd := exec.Command("yt-dlp", "-x", "--audio-format", "mp3", "-o", "theme.mp3", url)
	cmd.Dir = destPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("yt-dlp: %w\n%s", err, out)
	}
	return nil
}

func (gd GDrive) Download(url string, destPath string) error {
	resp, err := http.Get("https://drive.google.com/uc?export=download&id=" + strings.Split(url, "/")[5])
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(filepath.Join(destPath, "theme.mp3"))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func newDownloaders() map[string]Downloader {
	return map[string]Downloader{
		"youtube": Youtube{},
		"gdrive":  GDrive{},
	}
}
