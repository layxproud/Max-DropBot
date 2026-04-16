package bot

import (
	"context"
	"encoding/json"
	"era-dropbot/internal/attachments"
	"era-dropbot/internal/messages"

	"github.com/maxigo-bot/maxigo-client"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	client    *maxigo.Client
	processor *attachments.Processor
}

func NewHandler(cl *maxigo.Client, p *attachments.Processor) *Handler {
	return &Handler{cl, p}
}

func (h *Handler) handleCommand(ctx context.Context, chatID int64, text string) {
	switch text {

	case "/start":
		h.client.SendMessage(ctx, chatID, &maxigo.NewMessageBody{
			Text: maxigo.Some(string(messages.Start)),
		})

	case "/info":
		h.client.SendMessage(ctx, chatID, &maxigo.NewMessageBody{
			Text: maxigo.Some(string(messages.Info)),
		})

	default:
		h.client.SendMessage(ctx, chatID, &maxigo.NewMessageBody{
			Text: maxigo.Some(string(messages.UnknownCommand)),
		})
	}
}

func (h *Handler) Handle(ctx context.Context, raw json.RawMessage) {
	var base maxigo.Update
	_ = json.Unmarshal(raw, &base)

	switch base.UpdateType {
	case maxigo.UpdateMessageCreated:
		log.Info().Msg("New message received")

		var upd maxigo.MessageCreatedUpdate
		if err := json.Unmarshal(raw, &upd); err != nil {
			log.Error().Msgf("Unmarshal error: %s", err.Error())
			return
		}

		if upd.Message.Recipient.ChatID == nil {
			log.Error().Msg("ChatID is nil")
			return
		}
		chatID := *upd.Message.Recipient.ChatID

		// Обработчик команд
		text := ""
		if upd.Message.Body.Text != nil {
			text = *upd.Message.Body.Text
		}
		if text != "" && text[0] == '/' {
			h.handleCommand(ctx, chatID, text)
			return
		}

		// Обработчик вложений
		atts, err := upd.Message.Body.ParseAttachments()
		if err != nil {
			log.Error().Msgf("Parsing attachments failed: %s", err.Error())
			return
		}
		if len(atts) == 0 {
			return
		}
		h.processor.Process(ctx, chatID, atts)

	case maxigo.UpdateBotStarted:
		var upd maxigo.BotStartedUpdate
		if err := json.Unmarshal(raw, &upd); err != nil {
			log.Error().Msgf("Unmarshal error: %s", err.Error())
			return
		}

		_, err := h.client.SendMessage(ctx, upd.ChatID, &maxigo.NewMessageBody{
			Text: maxigo.Some(messages.Start),
		})

		if err != nil {
			log.Error().Msgf("Send message error: %s", err.Error())
			return
		}
	}
}
