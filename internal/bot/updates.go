package bot

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/maxigo-bot/maxigo-client"
)

func StartPolling(ctx context.Context, client *maxigo.Client, handler Handler) {
	var marker int64

	for {
		select {
		case <-ctx.Done():
			log.Println("Polling stopped")
			return
		default:
		}

		result, err := client.GetUpdates(ctx, maxigo.GetUpdatesOpts{
			Limit:   100,
			Timeout: 30,
			Marker:  marker,
		})

		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			log.Println("error:", err)
			select {
			case <-time.After(time.Second):
			case <-ctx.Done():
				return
			}

			continue
		}

		for _, raw := range result.Updates {
			handler.Handle(raw)
		}

		if result.Marker != nil {
			marker = *result.Marker
		}
	}
}
