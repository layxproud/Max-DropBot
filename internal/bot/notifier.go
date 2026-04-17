package bot

import (
	"context"
	"era-dropbot/internal/downloader"
	"era-dropbot/internal/messages"
	"fmt"

	"github.com/maxigo-bot/maxigo-client"
	"github.com/rs/zerolog/log"
)

func StartNotifier(ctx context.Context, client *maxigo.Client, results <-chan downloader.Result) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info().
					Msg("Notifier stopped")
				return

			case res, ok := <-results:
				if !ok {
					log.Info().
						Msg("Results channel closed")
					return
				}
				text := formatMessage(res)

				if text == "" {
					continue
				}
				_, err := client.SendMessage(ctx, res.ChatID, &maxigo.NewMessageBody{
					Text: maxigo.Some(text),
				})

				if err != nil {
					log.Error().
						Err(err).
						Msg("Send message error")
				}
			}
		}
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
