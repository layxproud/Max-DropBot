package main

import (
	"context"
	"era-dropbot/internal/attachments"
	"era-dropbot/internal/bot"
	"era-dropbot/internal/downloader"
	"era-dropbot/internal/storage"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	maxigo "github.com/maxigo-bot/maxigo-client"
)

func main() {
	_ = godotenv.Load()

	err := initFolders()
	if err != nil {
		log.Fatal(err)
	}

	client, err := maxigo.New(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	info, _ := client.GetBot(ctx)
	fmt.Printf("Бот: %s (ID: %d)\n", info.FirstName, info.UserID)

	dedup := storage.NewDeduplicator()
	pool := downloader.NewPool(4, dedup)
	processor := attachments.NewProcessor(pool)
	handler := bot.NewHandler(client, processor)

	bot.StartNotifier(ctx, client, pool.Results())
	bot.StartPolling(ctx, client, *handler)

	<-ctx.Done()

	pool.Close()
	log.Println("Shutdown complete")
}

func initFolders() error {
	err := os.MkdirAll("./downloads/pdfFiles", os.ModePerm)
	if err != nil {
		return err
	}

	err = os.MkdirAll("./downloads/pptFiles", os.ModePerm)
	if err != nil {
		return err
	}

	err = os.MkdirAll("./downloads/docFiles", os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
