package bot

import (
	"context"
	"encoding/json"
	"era-dropbot/internal/attachments"

	"github.com/maxigo-bot/maxigo-client"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	client    *maxigo.Client
	processor *attachments.Processor
	ctx       context.Context
}

func NewHandler(cl *maxigo.Client, p *attachments.Processor, ct context.Context) *Handler {
	return &Handler{cl, p, ct}
}

func (h *Handler) Handle(raw json.RawMessage) {
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

		atts, err := upd.Message.Body.ParseAttachments()
		if err != nil {
			log.Error().Msgf("Parsing attachments failed: %s", err.Error())
			return
		}

		if upd.Message.Recipient.ChatID == nil {
			log.Error().Msg("ChatID is nil")
			return
		}

		chatID := *upd.Message.Recipient.ChatID
		h.processor.Process(chatID, atts)

	case maxigo.UpdateBotStarted:
		var upd maxigo.BotStartedUpdate
		if err := json.Unmarshal(raw, &upd); err != nil {
			log.Error().Msgf("Unmarshal error: %s", err.Error())
			return
		}

		_, err := h.client.SendMessage(h.ctx, upd.ChatID, &maxigo.NewMessageBody{
			Text: maxigo.Some("Вас приветствует бот для загрузки файлов на сервер SAMPLE_NAME." +
				" Бот принимает файлы до 10 МБ следующих форматов:\n" +
				"1) PDF (.pdf)\n2) PowerPoint (.ppt, .pptx)\n" +
				"3) Microsoft Word: (.doc, .docx)\n" +
				"Для продолжения просто пришлите файл в этот чат."),
		})

		if err != nil {
			log.Error().Msgf("Send message error: %s", err.Error())
			return
		}
	}
}
