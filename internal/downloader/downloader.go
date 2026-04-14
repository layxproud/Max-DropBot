package downloader

import (
	"context"
	"era-dropbot/internal/storage"
	"era-dropbot/utils"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const MaxFileSize = 25 << 20 // 25 MB

type Downloader struct {
	client  *http.Client
	storage *storage.MinioStorage
}

func NewDownloader(storage *storage.MinioStorage) *Downloader {
	return &Downloader{
		client:  &http.Client{Timeout: 30 * time.Second},
		storage: storage,
	}
}

func backoff(ctx context.Context, attempt int) {
	base := time.Duration(1<<attempt) * time.Second
	jitter := time.Duration(rand.Int63n(int64(base / 2)))
	delay := base + jitter

	select {
	case <-time.After(delay):
	case <-ctx.Done():
	}
}

func (d *Downloader) Download(job Job) Result {
	log.Info().
		Str("file", job.Filename).
		Int64("chat_id", job.ChatID).
		Msg("start download")

	var resp *http.Response
	var err error

	// --- RETRY ---
	for i := 0; i < 3; i++ {
		select {
		case <-job.Ctx.Done():
			log.Info().Msg("downloader canceled")
			return Result{
				ChatID: job.ChatID,
				File:   job.Filename,
				Status: StatusInternalError,
			}
		default:
		}

		req, err := http.NewRequestWithContext(job.Ctx, "GET", job.URL, nil)
		if err != nil {
			log.Error().Err(err).Msg("create request failed")
			return Result{
				ChatID: job.ChatID,
				File:   job.Filename,
				Status: StatusInternalError,
			}
		}

		resp, err = d.client.Do(req)
		if err != nil {
			log.Warn().
				Err(err).
				Str("file", job.Filename).
				Int("attempt", i+1).
				Msg("retry download")

			backoff(job.Ctx, i)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			log.Warn().
				Str("file", job.Filename).
				Int("attempt", i+1).
				Msg("rate limited")

			resp.Body.Close()
			backoff(job.Ctx, i)
			continue
		}

		if resp.StatusCode >= 500 {
			log.Warn().
				Int("status", resp.StatusCode).
				Str("file", job.Filename).
				Int("attempt", i+1).
				Msg("server error")

			resp.Body.Close()
			backoff(job.Ctx, i)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			log.Error().
				Int("status", resp.StatusCode).
				Str("file", job.Filename).
				Msg("client error")

			return Result{
				ChatID: job.ChatID,
				File:   job.Filename,
				Status: StatusInternalError,
			}
		}

		break
	}

	if resp == nil {
		log.Error().Str("file", job.Filename).Msg("no response after retries")
		return Result{
			ChatID: job.ChatID,
			File:   job.Filename,
			Status: StatusInternalError,
		}
	}
	defer resp.Body.Close()

	// --- SIZE CHECK ---
	if resp.ContentLength > 0 && resp.ContentLength > MaxFileSize {
		log.Warn().
			Str("file", job.Filename).
			Int64("size", resp.ContentLength).
			Msg("file too large (header)")

		return Result{
			ChatID: job.ChatID,
			File:   job.Filename,
			Status: StatusTooLarge,
		}
	}

	// --- STREAM LIMIT ---
	limited := &io.LimitedReader{
		R: resp.Body,
		N: MaxFileSize,
	}

	// --- UNIQUE NAME ---
	safeName := utils.SanitizeFilename(job.Filename)
	objectName := fmt.Sprintf("%d_%s", time.Now().Unix(), safeName)

	// --- UPLOAD TO MINIO ---
	size := resp.ContentLength
	if size < 0 {
		size = -1
	}

	url, err := d.storage.Upload(
		job.Ctx,
		objectName,
		limited,
		size,
		resp.Header.Get("Content-Type"),
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("file", job.Filename).
			Msg("upload failed")

		return Result{
			ChatID: job.ChatID,
			File:   job.Filename,
			Status: StatusInternalError,
		}
	}

	// --- CHECK LIMIT OVERFLOW ---
	if limited.N <= 0 {
		log.Warn().
			Str("file", job.Filename).
			Msg("file exceeded limit during stream")

		return Result{
			ChatID: job.ChatID,
			File:   job.Filename,
			Status: StatusTooLarge,
		}
	}

	log.Info().
		Str("file", job.Filename).
		Str("url", url).
		Msg("file uploaded")

	return Result{
		ChatID: job.ChatID,
		File:   job.Filename,
		Status: StatusOK,
		URL:    url,
	}
}
