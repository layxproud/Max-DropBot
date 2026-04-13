package downloader

import (
	"era-dropbot/utils"
	"fmt"
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

func backoff(attempt int) {
	base := time.Duration(1<<attempt) * time.Second
	jitter := time.Duration(rand.Int63n(int64(base / 2)))
	time.Sleep(base + jitter)
}

func (d *Downloader) Download(job Job) Result {
	log.Info().Msgf("Downloading file %s in chat %d", job.Filename, job.ChatID)

	// Загрузка в несколько попыток
	var resp *http.Response
	var err error

	for i := 0; i < 3; i++ {
		select {
		case <-job.Ctx.Done():
			return Result{job.ChatID, job.Filename, "error", "Context was cancelled"}
		default:
		}

		req, err := http.NewRequestWithContext(job.Ctx, "GET", job.URL, nil)
		if err != nil {
			log.Error().
				Err(err).
				Msg("Create request error")
			return Result{job.ChatID, job.Filename, "error", err.Error()}
		}

		resp, err = d.client.Do(req)
		if err != nil {
			log.Warn().
				Err(err).
				Msgf("Download retry %d file=%s",
					i+1, job.Filename)

			backoff(i)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			log.Warn().
				Msgf("Rate limited retry %d file=%s",
					i+1, job.Filename)

			resp.Body.Close()
			backoff(i)
			continue
		}

		if resp.StatusCode >= 500 {
			log.Warn().
				Msgf("Server error %d retry %d file=%s",
					resp.StatusCode, i+1, job.Filename)

			resp.Body.Close()
			backoff(i)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			log.Error().
				Msgf("Client error %d file=%s",
					resp.StatusCode, job.Filename)
			return Result{job.ChatID, job.Filename, "error", fmt.Sprintf("Status %d", resp.StatusCode)}
		}

		break
	}
	defer resp.Body.Close()

	// Проверяем размер
	if resp.ContentLength > MaxFileSize {
		log.Error().
			Msgf("File=%s size=%d is rejected, too large",
				job.Filename, resp.ContentLength)
		return Result{job.ChatID, job.Filename, "too_large", "Файл слишком большой"}
	}

	// Создаем пути
	safeName := utils.SanitizeFilename(job.Filename)
	finalPath := job.Path + safeName
	tmpPath := finalPath + ".tmp"

	// Создаем временный файл
	file, err := os.Create(tmpPath)
	if err != nil {
		log.Error().
			Err(err).
			Msgf("Couldn't create temp file %s", job.Filename)
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}

	// Если что-то пошло не так - удалим tmp
	defer func() {
		file.Close()
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
			Msgf("Couldn't write file %s", job.Filename)
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}

	// файл оборван
	if resp.ContentLength > 0 && written != resp.ContentLength {
		log.Error().
			Msgf("Incomplete file %s | Written=%d / Expected=%d",
				job.Filename, written, resp.ContentLength)

		return Result{job.ChatID, job.Filename, "error", "Файл скачан не полностью"}
	}

	// превышен лимит
	if limited.N <= 0 {
		log.Error().
			Msgf("File=%s size=%d is rejected, too large",
				job.Filename, resp.ContentLength)
		return Result{job.ChatID, job.Filename, "too_large", "Файл слишком большой"}
	}

	// Очистка буфера и завершение работы
	if err := file.Sync(); err != nil {
		log.Error().
			Err(err).
			Msgf("Failed syncing file %s", job.Filename)
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}

	if err := file.Close(); err != nil {
		log.Error().
			Err(err).
			Msgf("Couldn't close file %s", job.Filename)
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		log.Error().
			Err(err).
			Msgf("Couldn't rename file %s", job.Filename)
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}

	log.Info().
		Msgf("Saved file %s", job.Filename)
	return Result{job.ChatID, job.Filename, "ok", "Файл успешно сохранен"}
}
