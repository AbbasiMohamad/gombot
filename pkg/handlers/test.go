package handlers

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"gombot/pkg/entities"
	"gombot/pkg/repositories"
)

func TestHandler(ctx context.Context, b *bot.Bot, update *models.Update) {

	job := entities.Job{
		JobId:        uuid.New(),
		ChatId:       update.Message.Chat.ID,
		MessageId:    update.Message.ID,
		Status:       entities.Requested,
		Applications: make([]entities.Application, 0), // TODO: what is zero?!
	}

	repositories.InsertJob(&job)

	SendMessage(b, ctx, update.Message.Chat.ID, fmt.Sprintf("work done for job %d", job.ID))
}
