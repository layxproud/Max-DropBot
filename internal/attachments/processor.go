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
	".pdf":  "./downloads/pdfFiles/",
	".doc":  "./downloads/docFiles/",
	".docx": "./downloads/docFiles/",
	".ppt":  "./downloads/pptFiles/",
	".pptx": "./downloads/pptFiles/",
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
