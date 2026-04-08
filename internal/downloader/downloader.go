package downloader

import (
	"crypto/sha256"
	"encoding/hex"
	"era-dropbot/utils"
	"io"
	"net/http"
	"os"
	"time"
)

const MaxFileSize = 10 << 20 // 10 MB

type Deduplicator interface {
	Seen(hash string) bool
	Store(hash string)
}

type Downloader struct {
	client *http.Client
	dedup  Deduplicator
}

func NewDownloader(d Deduplicator) *Downloader {
	return &Downloader{
		client: &http.Client{Timeout: 30 * time.Second},
		dedup:  d,
	}
}

func (d *Downloader) Download(job Job) Result {
	resp, err := d.client.Get(job.URL)
	if err != nil {
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}
	defer resp.Body.Close()

	if resp.ContentLength > MaxFileSize {
		return Result{job.ChatID, job.Filename, "too_large", "Файл слишком большой"}
	}

	safeName := utils.SanitizeFilename(job.Filename)
	path := job.Path + safeName

	file, err := os.Create(path)
	if err != nil {
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}
	defer file.Close()

	hasher := sha256.New()
	tee := io.TeeReader(resp.Body, hasher)

	limited := &io.LimitedReader{
		R: tee,
		N: MaxFileSize,
	}

	_, err = io.Copy(file, limited)
	if err != nil {
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}

	hash := hex.EncodeToString(hasher.Sum(nil))

	if d.dedup.Seen(hash) {
		_ = os.Remove(job.Path + safeName)
		return Result{job.ChatID, job.Filename, "duplicate", "Файл уже загружался"}
	}

	d.dedup.Store(hash)
	return Result{job.ChatID, job.Filename, "ok", "Файл успешно сохранен"}
}
