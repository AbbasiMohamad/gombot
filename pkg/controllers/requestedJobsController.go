package controllers

import (
	"context"
	"github.com/go-telegram/bot"
	"gombot/pkg/domain/entities"
	"gombot/pkg/handlers"
	"gombot/pkg/repositories"
	"log"
)

func ExecuteMonitoringOfRequestedJobs(ctx context.Context, b *bot.Bot) {
	for {
		jobs := repositories.GetRequestedJobs()
		if len(jobs) == 0 {
			sleep()
		} else {
			for i := range jobs {
				handleRequestedJob(jobs[i], ctx, b)
			}
		}
	}
}

// handleRequestedJob send 'Request Message' to user for approve. then make Job's status to 'NeedToApproved' and
// Application's status to 'Pending'
func handleRequestedJob(job *entities.Job, ctx context.Context, b *bot.Bot) {
	messageId := sendRequestMessage(job, ctx, b)
	err := job.SetRequestMessageID(messageId)
	if err != nil {
		log.Printf("can not set request message id: %v", err)
	}
	repositories.UpdateJob(job)
}

func sendRequestMessage(job *entities.Job, ctx context.Context, b *bot.Bot) int {
	message := handlers.MakeRequestMessageContent(job)
	msg := handlers.SendMessageWithInlineKeyboardMarkup(b, ctx, job.ChatId, message)
	return msg.ID
}
