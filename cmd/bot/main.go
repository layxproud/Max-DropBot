package main

import (
	"context"
	"era-dropbot/internal/attachments"
	"era-dropbot/internal/bot"
	"era-dropbot/internal/downloader"
	"era-dropbot/internal/shortener"
	"era-dropbot/internal/storage"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	maxigo "github.com/maxigo-bot/maxigo-client"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func main() {
	_ = godotenv.Load()

	// err := initFolders()
	// if err != nil {
	// 	log.Fatal().Msgf("Init folders error: %s", err.Error())
	// }

	client, err := maxigo.New(
		os.Getenv("BOT_TOKEN"),
		maxigo.WithRetry(),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("create bot error")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info().Msg("shutting down...")
		cancel()
	}()

	info, err := client.GetBot(ctx)
	if err != nil {
		log.Error().Err(err).Msg("couldn't get bot info")
	}
	log.Info().
		Str("bot_name", info.FirstName).
		Int64("bot_id", info.UserID).
		Msg("bot information")

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	short := shortener.NewShortener(rdb, os.Getenv("SHORT_PREFIX"))

	go http.ListenAndServe(":8080", shortener.NewServer(short))

	minioStorage, err := storage.NewMinio(
		os.Getenv("MINIO_ENDPOINT"),
		os.Getenv("MINIO_ROOT_USER"),
		os.Getenv("MINIO_ROOT_PASSWORD"),
		"files",
		os.Getenv("MINIO_PUBLIC_ENDPOINT"),
		short,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init minio")
	}

	pool := downloader.NewPool(4, minioStorage)
	processor := attachments.NewProcessor(pool)
	handler := bot.NewHandler(client, processor)

	bot.StartNotifier(ctx, client, pool.Results())
	bot.StartPolling(ctx, client, *handler)

	<-ctx.Done()

	pool.Close()
	log.Info().Msg("shutdown complete!")
}

func initFolders() error {
	dirs := []string{"PDF", "Word", "PowerPoint", "Video", "Audio", "Image"}

	for _, d := range dirs {
		if err := os.MkdirAll("./downloads/"+d, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
