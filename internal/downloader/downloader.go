package downloader

import (
	"context"
	"era-dropbot/utils"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

const MaxFileSize = 25 << 20 // 25 MB

type Downloader struct {
	client *http.Client
}

func NewDownloader() *Downloader {
	return &Downloader{
		client: &http.Client{Timeout: 30 * time.Second},
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
		Msgf("Downloading file %s in chat %d", job.Filename, job.ChatID)

	// Загрузка в несколько попыток
	var resp *http.Response
	var err error

	for i := 0; i < 3; i++ {
		select {
		case <-job.Ctx.Done():
			log.Info().
				Msg("Downloader stopped")
			return Result{
				job.ChatID,
				job.Filename,
				StatusInternalError,
			}
		default:
		}

		req, err := http.NewRequestWithContext(job.Ctx, "GET", job.URL, nil)
		if err != nil {
			log.Error().
				Err(err).
				Msg("Create request error")
			return Result{
				job.ChatID,
				job.Filename,
				StatusInternalError,
			}
		}

		resp, err = d.client.Do(req)
		if err != nil {
			log.Warn().
				Err(err).
				Str("File", job.Filename).
				Int("Attempt", i+1).
				Msg("Download retry")

			backoff(job.Ctx, i)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			log.Warn().
				Str("File", job.Filename).
				Int("Attempt", i+1).
				Msg("Rate limited")

			resp.Body.Close()
			backoff(job.Ctx, i)
			continue
		}

		if resp.StatusCode >= 500 {
			log.Warn().
				Int("Code", resp.StatusCode).
				Str("File", job.Filename).
				Int("Attempt", i+1).
				Msg("Server error")

			resp.Body.Close()
			backoff(job.Ctx, i)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			log.Error().
				Int("Code", resp.StatusCode).
				Str("File", job.Filename).
				Msg("Client error")
			return Result{
				job.ChatID,
				job.Filename,
				StatusInternalError,
			}
		}

		break
	}
	if resp == nil {
		log.Error().
			Str("File", job.Filename).
			Msg("No response")

		return Result{
			job.ChatID,
			job.Filename,
			StatusInternalError,
		}
	}

	defer resp.Body.Close()

	// Проверяем размер
	if resp.ContentLength > 0 && resp.ContentLength > MaxFileSize {
		log.Error().
			Str("File", job.Filename).
			Int("Size", int(resp.ContentLength)).
			Msg("Too large")
		return Result{
			job.ChatID,
			job.Filename,
			StatusTooLarge,
		}
	}

	// Создаем пути
	safeName := utils.SanitizeFilename(job.Filename)
	finalPath := job.Path + safeName
	tmpPath := finalPath + ".tmp"

	// Создаем временный файл
	file, err := os.Create(tmpPath)
	if err != nil {
		log.Error().
			Str("File", job.Filename).
			Err(err).
			Msg("Create file error")
		return Result{
			job.ChatID,
			job.Filename,
			StatusInternalError,
		}
	}

	// Если что-то пошло не так - удалим tmp
	defer func() {
		if file != nil {
			file.Close()
		}
		_ = os.Remove(tmpPath)
	}()

	limited := &io.LimitedReader{
		R: resp.Body,
		N: MaxFileSize,
	}

	// Копируем данные в файл
	written, err := io.Copy(file, limited)
	if err != nil {
		log.Error().
			Err(err).
			Str("File", job.Filename).
			Msg("Write file error")
		return Result{
			job.ChatID,
			job.Filename,
			StatusInternalError,
		}
	}

	// файл оборван
	if resp.ContentLength > 0 && written != resp.ContentLength {
		log.Error().
			Str("File", job.Filename).
			Msg("Incomplete download")
		return Result{
			job.ChatID,
			job.Filename,
			StatusInternalError,
		}
	}

	// превышен лимит
	if limited.N <= 0 {
		log.Error().
			Str("File", job.Filename).
			Int("Size", int(resp.ContentLength)).
			Msg("Too large")
		return Result{job.ChatID, job.Filename, StatusTooLarge}
	}

	// Очистка буфера и завершение работы
	if err := file.Sync(); err != nil {
		log.Error().
			Err(err).
			Str("File", job.Filename).
			Msg("Flush error")
		return Result{
			job.ChatID,
			job.Filename,
			StatusInternalError,
		}
	}

	if err := file.Close(); err != nil {
		log.Error().
			Err(err).
			Str("File", job.Filename).
			Msg("Close file error")
		return Result{
			job.ChatID,
			job.Filename,
			StatusInternalError,
		}
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		log.Error().
			Err(err).
			Str("File", job.Filename).
			Msg("Rename file error")
		return Result{
			job.ChatID,
			job.Filename,
			StatusInternalError,
		}
	}
	file = nil

	// Успех
	log.Info().
		Str("File", job.Filename).
		Msg("File saved")
	return Result{
		job.ChatID,
		job.Filename,
		StatusOK,
	}
}
