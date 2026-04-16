package attachments

import (
	"context"
	"era-dropbot/internal/downloader"
	"strings"

	"github.com/maxigo-bot/maxigo-client"
	"github.com/rs/zerolog/log"
)

type Processor struct {
	pool *downloader.Pool
}

func NewProcessor(pool *downloader.Pool) *Processor {
	return &Processor{pool: pool}
}

var routes = map[string]string{
	".pdf":  "pdf",
	".doc":  "word",
	".docx": "word",
	".ppt":  "ppt",
	".pptx": "ppt",
	".mp4":  "video",
	".mov":  "video",
	".mkv":  "video",
	".avi":  "video",
	".wmv":  "video",
	".mp3":  "audio",
	".wav":  "audio",
	".aac":  "audio",
	".m4a":  "audio",
	".flac": "audio",
	".png":  "image",
	".jpg":  "image",
	".jpeg": "image",
	".gif":  "image",
	".tiff": "image",
	".svg":  "image",
}

func (p *Processor) Process(ctx context.Context, chatID int64, atts []maxigo.Attachment) {
	for _, att := range atts {
		switch a := att.(type) {
		case *maxigo.FileAttachment:
			p.handleFile(ctx, chatID, a)
		default:
			log.Warn().Str("type", a.GetType()).Msg("unsupported type of attachment")
			p.pool.SendResult(downloader.Result{
				ChatID: chatID,
				File:   "unknown",
				Status: downloader.StatusUnsupported,
			})
		}
	}
}

func (p *Processor) handleFile(ctx context.Context, chatID int64, file *maxigo.FileAttachment) {
	name := strings.ToLower(file.Filename)

	for ext, path := range routes {
		if strings.HasSuffix(name, ext) {
			p.pool.Submit(downloader.Job{
				URL:      file.Payload.URL,
				Filename: file.Filename,
				Path:     path,
				ChatID:   chatID,
				Ctx:      ctx,
			})
			return
		}
	}

	log.Warn().Str("file", file.Filename).Msg("unsupported file extension")
	p.pool.SendResult(downloader.Result{
		ChatID: chatID,
		File:   file.Filename,
		Status: downloader.StatusUnsupported,
	})
}
