package bot

import (
	"context"
	"errors"
	"time"

	"github.com/maxigo-bot/maxigo-client"
	"github.com/rs/zerolog/log"
)

func StartPolling(ctx context.Context, client *maxigo.Client, handler Handler) {
	var marker int64
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("polling cancelled")
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
			var e *maxigo.Error
			if errors.As(err, &e) && e.Kind == maxigo.ErrTimeout {
				continue
			}
			log.Error().Err(err).Msg("polling error")
			select {
			case <-time.After(time.Second):
			case <-ctx.Done():
				return
			}
			continue
		}

		for _, raw := range result.Updates {
			handler.Handle(ctx, raw)
		}
		if result.Marker != nil {
			marker = *result.Marker
		}
	}
}
