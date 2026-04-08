package bot

import (
	"context"
	"era-dropbot/internal/downloader"
	"log"

	"github.com/maxigo-bot/maxigo-client"
)

func StartNotifier(ctx context.Context, client *maxigo.Client, results <-chan downloader.Result) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Notifier stopped")
				return

			case res := <-results:
				text := formatMessage(res)

				_, err := client.SendMessage(ctx, res.ChatID, &maxigo.NewMessageBody{
					Text: maxigo.Some(text),
				})

				if err != nil {
					log.Println("send error:", err)
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
