package downloader

import (
	"crypto/sha256"
	"encoding/hex"
	"era-dropbot/utils"
	"io"
	"log"
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
	var resp *http.Response
	var err error

	log.Printf("Downloading file: %s from chat %d", job.Filename, job.ChatID)
	for i := 0; i < 3; i++ {
		resp, err = d.client.Get(job.URL)
		if err == nil {
			break
		}

		time.Sleep(time.Duration(1<<i) * time.Second)
	}

	if err != nil {
		log.Printf("Download failed: %s error: %v", job.Filename, err)
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}
	defer resp.Body.Close()

	if resp.ContentLength > MaxFileSize {
		log.Printf("File %s is too large", job.Filename)
		return Result{job.ChatID, job.Filename, "too_large", "Файл слишком большой"}
	}

	safeName := utils.SanitizeFilename(job.Filename)
	path := job.Path + safeName

	file, err := os.Create(path)
	if err != nil {
		log.Printf("Download failed: %s error: %v", job.Filename, err)
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
		log.Printf("Download failed: %s error: %v", job.Filename, err)
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}

	hash := hex.EncodeToString(hasher.Sum(nil))

	if d.dedup.Seen(hash) {
		log.Printf("Duplicate file skipped: %s", job.Filename)
		_ = os.Remove(job.Path + safeName)
		return Result{job.ChatID, job.Filename, "duplicate", "Файл уже загружался"}
	}

	d.dedup.Store(hash)
	log.Printf("Saved file: %s", job.Filename)
	return Result{job.ChatID, job.Filename, "ok", "Файл успешно сохранен"}
}
