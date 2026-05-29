package bot

import (
	"context"
	"fmt"
	"max-dropbot/internal/downloader"
	"max-dropbot/internal/messages"

	"github.com/maxigo-bot/maxigo-client"
	"github.com/rs/zerolog/log"
)

func StartNotifier(client *maxigo.Client, results <-chan downloader.Result) {
	go func() {
		for res := range results {
			text := formatMessage(res)
			if text == "" {
				continue
			}
			if _, err := client.SendMessage(context.Background(), res.ChatID, &maxigo.NewMessageBody{
				Text: maxigo.Some(text),
			}); err != nil {
				log.Error().Err(err).Msg("send message error")
			}
		}
		log.Info().Msg("notifier done")
	}()
}

func formatMessage(r downloader.Result) string {
	switch r.Status {
	case downloader.StatusOK:
		return fmt.Sprintf(messages.FileUploaded, r.File, r.URL)

	case downloader.StatusTooLarge:
		return fmt.Sprintf(messages.FileTooLarge, r.File)

	case downloader.StatusUnsupported:
		return fmt.Sprintf(messages.UnsupportedFormat, r.File)

	case downloader.StatusInternalError:
		return fmt.Sprintf(messages.InternalError, r.File)
	}

	return ""
}
