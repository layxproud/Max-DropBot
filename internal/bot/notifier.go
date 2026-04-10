package bot

import (
	"context"
	"era-dropbot/internal/downloader"

	"github.com/maxigo-bot/maxigo-client"
	"github.com/rs/zerolog/log"
)

func StartNotifier(ctx context.Context, client *maxigo.Client, results <-chan downloader.Result) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("Notifier stopped")
				return

			case res, ok := <-results:
				if !ok {
					log.Info().Msg("Results channel closed")
					return
				}
				text := formatMessage(res)

				_, err := client.SendMessage(ctx, res.ChatID, &maxigo.NewMessageBody{
					Text: maxigo.Some(text),
				})

				if err != nil {
					log.Error().Msgf("Send message error: %s", err.Error())
				}
			}
		}
	}()
}

func formatMessage(r downloader.Result) string {
	switch r.Status {
	case "ok":
		return "✅ " + r.File + " сохранён"
	case "too_large":
		return "❌ " + r.File + " слишком большой"
	case "unsupported_type", "unsupported_format":
		return "🚫 " + r.Message
	default:
		return "❌ Ошибка: " + r.Message
	}
}
