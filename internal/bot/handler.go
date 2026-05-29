package bot

import (
	"context"
	"encoding/json"
	"max-dropbot/internal/attachments"
	"max-dropbot/internal/messages"

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
		if _, err := h.client.SendMessage(ctx, chatID, &maxigo.NewMessageBody{
			Text: maxigo.Some(string(messages.Start)),
		}); err != nil {
			log.Error().Err(err).Int64("chatID", chatID).Msg("send message error")
		}

	case "/info":
		if _, err := h.client.SendMessage(ctx, chatID, &maxigo.NewMessageBody{
			Text: maxigo.Some(string(messages.Info)),
		}); err != nil {
			log.Error().Err(err).Int64("chatID", chatID).Msg("send message error")
		}

	default:
		if _, err := h.client.SendMessage(ctx, chatID, &maxigo.NewMessageBody{
			Text: maxigo.Some(string(messages.UnknownCommand)),
		}); err != nil {
			log.Error().Err(err).Int64("chatID", chatID).Msg("send message error")
		}
	}
}

func (h *Handler) Handle(ctx context.Context, raw json.RawMessage) {
	var base maxigo.Update
	if err := json.Unmarshal(raw, &base); err != nil {
		log.Error().Err(err).Msg("unmarshal base update error")
		return
	}

	switch base.UpdateType {
	case maxigo.UpdateMessageCreated:
		log.Info().Msg("new message received")

		var upd maxigo.MessageCreatedUpdate
		if err := json.Unmarshal(raw, &upd); err != nil {
			log.Error().Err(err).Msg("unmarshal error")
			return
		}

		if upd.Message.Recipient.ChatID == nil {
			log.Error().Msg("chatID is nil")
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
			log.Error().Err(err).Msg("parsing attachments failed")
			return
		}
		if len(atts) == 0 {
			return
		}
		h.processor.Process(ctx, chatID, atts)

	case maxigo.UpdateBotStarted:
		var upd maxigo.BotStartedUpdate
		if err := json.Unmarshal(raw, &upd); err != nil {
			log.Error().Err(err).Msgf("unmarshal error")
			return
		}

		if _, err := h.client.SendMessage(ctx, upd.ChatID, &maxigo.NewMessageBody{
			Text: maxigo.Some(messages.Start),
		}); err != nil {
			log.Error().Err(err).Int64("chatID", upd.ChatID).Msg("send message error")
		}
	}
}
