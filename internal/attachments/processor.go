package attachments

import (
	"era-dropbot/internal/downloader"
	"strings"

	"github.com/maxigo-bot/maxigo-client"
)

type Processor struct {
	pool *downloader.Pool
}

func NewProcessor(pool *downloader.Pool) *Processor {
	return &Processor{pool: pool}
}

var routes = map[string]string{
	".pdf":  "./downloads/PDF/",
	".doc":  "./downloads/Word/",
	".docx": "./downloads/Word/",
	".ppt":  "./downloads/PowerPoint/",
	".pptx": "./downloads/PowerPoint/",
	".mp4":  "./downloads/Video/",
	".mov":  "./downloads/Video/",
	".mkv":  "./downloads/Video/",
	".avi":  "./downloads/Video/",
	".wmv":  "./downloads/Video/",
	".mp3":  "./downloads/Audio/",
	".wav":  "./downloads/Audio/",
	".aac":  "./downloads/Audio/",
	".m4a":  "./downloads/Audio/",
	".flac": "./downloads/Audio/",
	".png":  "./downloads/Image/",
	".jpg":  "./downloads/Image/",
	".jpeg": "./downloads/Image/",
	".gif":  "./downloads/Image/",
	".tiff": "./downloads/Image/",
	".svg":  "./downloads/Image/",
}

func (p *Processor) Process(chatID int64, atts []maxigo.Attachment) {
	for _, att := range atts {
		switch a := att.(type) {
		case *maxigo.FileAttachment:
			p.handleFile(chatID, a)
		default:
			p.pool.SendResult(downloader.Result{
				ChatID:  chatID,
				File:    "unknown",
				Status:  "unsupported_type",
				Message: "Неподдерживаемый тип вложения",
			})
		}
	}
}

func (p *Processor) handleFile(chatID int64, file *maxigo.FileAttachment) {
	name := strings.ToLower(file.Filename)

	for ext, path := range routes {
		if strings.HasSuffix(name, ext) {
			p.pool.Submit(downloader.Job{
				URL:      file.Payload.URL,
				Filename: file.Filename,
				Path:     path,
				ChatID:   chatID,
			})
			return
		}
	}

	p.pool.SendResult(downloader.Result{
		ChatID:  chatID,
		File:    file.Filename,
		Status:  "unsupported_format",
		Message: "Неподдерживаемый формат файла",
	})
}
