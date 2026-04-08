package bot

import (
	"context"
	"encoding/json"
	"era-dropbot/internal/attachments"
	"log"

	"github.com/maxigo-bot/maxigo-client"
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
		log.Println("Message received")
		var upd maxigo.MessageCreatedUpdate
		_ = json.Unmarshal(raw, &upd)

		atts, err := upd.Message.Body.ParseAttachments()
		if err != nil {
			log.Println(err)
			return
		}

		h.processor.Process(*upd.Message.Recipient.ChatID, atts)

	case maxigo.UpdateBotStarted:
		var upd maxigo.BotStartedUpdate
		_ = json.Unmarshal(raw, &upd)

		_, err := h.client.SendMessage(h.ctx, upd.ChatID, &maxigo.NewMessageBody{
			Text: maxigo.Some("Вас приветствует бот для загрузки файлов на сервер SAMPLE_NAME." +
				" Бот принимает файлы до 10 МБ следующих форматов:\n" +
				"\n1) PDF (.pdf)\n2) PowerPoint (.ppt, .pptx)\n" +
				"3) Microsoft Word: (.doc, .docx)\n" +
				"Для продолжения просто пришлите файл в этот чат."),
		})

		if err != nil {
			log.Println("send error:", err)
		}
	}
}
