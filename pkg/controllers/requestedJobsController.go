package controllers

import (
	"context"
	"github.com/go-telegram/bot"
	"gombot/pkg/domain/entities"
	"gombot/pkg/handlers"
	"gombot/pkg/repositories"
)

func ExecuteMonitoringOfRequestedJobs(ctx context.Context, b *bot.Bot) {
	for {

		jobs := repositories.GetRequestedJobs()
		if len(jobs) == 0 {
			sleep()
			continue
		}

		if len(jobs) > 0 {
			for i, _ := range jobs {
				if jobs[i].Status == entities.Requested {
					messageId := handlers.SendMessageForApprove(*jobs[i], ctx, b)
					for j, _ := range jobs[i].Applications {
						jobs[i].Applications[j].Status = entities.Pending
					}
					jobs[i].Status = entities.NeedToApproved
					//important: update message id of job
					if jobs[i].RequestMessageID == 0 {
						jobs[i].RequestMessageID = messageId
					}
					repositories.UpdateJob(jobs[i])
				}
			}
			continue
		}
	}
}
