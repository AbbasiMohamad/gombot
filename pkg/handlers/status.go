package handlers

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	model "gombot/pkg/domain/entities"
)

func StatusHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	for _, applications := range model.Queue {
		for _, app := range applications.Applications {
			SendMessage(b, ctx, update.Message.Chat.ID, app.Name)
		}
	}
}
