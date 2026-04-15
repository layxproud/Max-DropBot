package bot

import (
	"context"
	"era-dropbot/internal/downloader"
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
					return
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
		return fmt.Sprintf("✅ %s загружен!\n%s", r.File, r.URL)

	case downloader.StatusTooLarge:
		return fmt.Sprintf("❌ %s слишком большой!\nМаксимальный размер 25 МБ.", r.File)

	case downloader.StatusUnsupported:
		return fmt.Sprintf("❌ Не удалось скачать файл %s\n"+
			"Для получения справки по принимаемым форматам напишите /info", r.File)

	case downloader.StatusInternalError:
		return fmt.Sprintf("❌ Не удалось скачать файл %s из-за внутренней ошибки.", r.File)
	}

	return ""
}
