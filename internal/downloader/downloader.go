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
	log.Printf("[DOWNLOAD] start file=%s chat=%d", job.Filename, job.ChatID)

	// Загрузка в несколько попыток
	var resp *http.Response
	var err error

	for i := 0; i < 3; i++ {
		resp, err = d.client.Get(job.URL)
		if err == nil {
			break
		}

		log.Printf("[DOWNLOAD] retry=%d file=%s error=%v", i+1, job.Filename, err)

		time.Sleep(time.Duration(1<<i) * time.Second)
	}

	if err != nil {
		log.Printf("[ERROR] download failed file=%s error=%v", job.Filename, err)
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}
	defer resp.Body.Close()

	// Проверяем размер
	if resp.ContentLength > MaxFileSize {
		log.Printf("[REJECT] too large file=%s size=%d", job.Filename, resp.ContentLength)
		return Result{job.ChatID, job.Filename, "too_large", "Файл слишком большой"}
	}

	// Создаем пути
	safeName := utils.SanitizeFilename(job.Filename)
	finalPath := job.Path + safeName
	tmpPath := finalPath + ".tmp"

	// Создаем временный файл
	file, err := os.Create(tmpPath)
	if err != nil {
		log.Printf("[ERROR] create tmp file failed file=%s error=%v", job.Filename, err)
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}

	// Если что-то пошло не так - удалим tmp
	defer func() {
		file.Close()
		_ = os.Remove(tmpPath)
	}()

	hasher := sha256.New()
	tee := io.TeeReader(resp.Body, hasher)

	limited := &io.LimitedReader{
		R: tee,
		N: MaxFileSize,
	}

	// Копируем данные в файл
	written, err := io.Copy(file, limited)
	if err != nil {
		log.Printf("[ERROR] write failed file=%s error=%v", job.Filename, err)
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}

	// файл оборван
	if resp.ContentLength > 0 && written != resp.ContentLength {
		log.Printf("[ERROR] incomplete file=%s written=%d expected=%d",
			job.Filename, written, resp.ContentLength)

		return Result{job.ChatID, job.Filename, "error", "Файл скачан не полностью"}
	}

	// превышен лимит
	if limited.N <= 0 {
		log.Printf("[REJECT] exceeded limit file=%s", job.Filename)
		return Result{job.ChatID, job.Filename, "too_large", "Файл превышает лимит"}
	}

	// Проверяем хеш
	hash := hex.EncodeToString(hasher.Sum(nil))

	if d.dedup.Seen(hash) {
		log.Printf("[DUPLICATE] file=%s hash=%s", job.Filename, hash)
		return Result{job.ChatID, job.Filename, "duplicate", "Файл уже загружался"}
	}

	// Очистка буфера и завершение работы
	if err := file.Sync(); err != nil {
		log.Printf("[ERROR] sync failed file=%s error=%v", job.Filename, err)
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}

	if err := file.Close(); err != nil {
		log.Printf("[ERROR] close failed file=%s error=%v", job.Filename, err)
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		log.Printf("[ERROR] rename failed file=%s error=%v", job.Filename, err)
		return Result{job.ChatID, job.Filename, "error", err.Error()}
	}

	// Сохранение хеша
	d.dedup.Store(hash)

	log.Printf("[SUCCESS] saved file=%s size=%d", job.Filename, written)

	return Result{job.ChatID, job.Filename, "ok", "Файл успешно сохранен"}
}
