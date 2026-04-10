package downloader

import (
	"era-dropbot/utils"
	"io"
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

func (d *Downloader) Download(job Job) Result {
	log.Info().Msgf("Downloading file %s in chat %d", job.Filename, job.ChatID)

	// Загрузка в несколько попыток
	var resp *http.Response
	var err error

	for i := range 3 {
		resp, err = d.client.Get(job.URL)
		if err == nil {
			break
		}
		log.Info().Msgf("Download retry %d file %s | Error: %s", i+1, job.Filename, err.Error())
		time.Sleep(time.Duration(1<<i) * time.Second)
	}

	if err != nil {
		log.Error().Msgf("Failed downloading file %s | Error: %s", job.Filename, err.Error())
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}
	defer resp.Body.Close()

	// Проверяем размер
	if resp.ContentLength > MaxFileSize {
		log.Error().Msgf("File=%s size=%d is rejected, too large", job.Filename, resp.ContentLength)
		return Result{job.ChatID, job.Filename, "too_large", "Файл слишком большой"}
	}

	// Создаем пути
	safeName := utils.SanitizeFilename(job.Filename)
	finalPath := job.Path + safeName
	tmpPath := finalPath + ".tmp"

	// Создаем временный файл
	file, err := os.Create(tmpPath)
	if err != nil {
		log.Error().Msgf("Couldn't create temp file %s | Error: %s", job.Filename, err.Error())
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
		log.Error().Msgf("Failed writing file %s | Error: %s", job.Filename, err.Error())
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}

	// файл оборван
	if resp.ContentLength > 0 && written != resp.ContentLength {
		log.Printf("Incomplete file %s | Written=%d / Expected=%d",
			job.Filename, written, resp.ContentLength)

		return Result{job.ChatID, job.Filename, "error", "Файл скачан не полностью"}
	}

	// превышен лимит
	if limited.N <= 0 {
		log.Error().Msgf("File=%s size=%d is rejected, too large", job.Filename, resp.ContentLength)
		return Result{job.ChatID, job.Filename, "too_large", "Файл слишком большой"}
	}

	// Очистка буфера и завершение работы
	if err := file.Sync(); err != nil {
		log.Error().Msgf("Failed syncing file %s | Error: %s", job.Filename, err.Error())
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}

	if err := file.Close(); err != nil {
		log.Error().Msgf("Couldn't close file %s | Error: %v", job.Filename, err.Error())
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		log.Printf("Couldn't rename file %s | Error: %s", job.Filename, err.Error())
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}

	log.Info().Msgf("Saved file %s", job.Filename)

	return Result{job.ChatID, job.Filename, "ok", "Файл успешно сохранен"}
}
