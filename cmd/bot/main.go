package main

import (
	"context"
	"era-dropbot/internal/attachments"
	"era-dropbot/internal/bot"
	"era-dropbot/internal/downloader"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	maxigo "github.com/maxigo-bot/maxigo-client"
	"github.com/rs/zerolog/log"
)

func main() {
	_ = godotenv.Load()

	err := initFolders()
	if err != nil {
		log.Fatal().Msgf("Init folders error: %s", err.Error())
	}

	client, err := maxigo.New(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatal().Msgf("Create bot error: %s", err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info().Msg("Shutting down...")
		cancel()
	}()

	info, err := client.GetBot(ctx)
	if err != nil {
		log.Error().Msgf("Couldn't get bot info: %s", err.Error())
	}
	log.Info().Msgf("Бот: %s (ID: %d)\n", info.FirstName, info.UserID)

	pool := downloader.NewPool(4)
	processor := attachments.NewProcessor(pool)
	handler := bot.NewHandler(client, processor, ctx)

	bot.StartNotifier(ctx, client, pool.Results())
	bot.StartPolling(ctx, client, *handler)

	<-ctx.Done()

	pool.Close()
	log.Info().Msg("Shutdown complete!")
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
