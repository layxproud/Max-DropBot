package bot

import (
	"encoding/json"
	"era-dropbot/internal/attachments"
	"log"

	"github.com/maxigo-bot/maxigo-client"
)

type Handler struct {
	client    *maxigo.Client
	processor *attachments.Processor
}

func NewHandler(c *maxigo.Client, p *attachments.Processor) *Handler {
	return &Handler{c, p}
}

func (h *Handler) Handle(raw json.RawMessage) {
	var base maxigo.Update
	_ = json.Unmarshal(raw, &base)

	switch base.UpdateType {
	case maxigo.UpdateMessageCreated:
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
	}
}
