package controllers

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"gombot/pkg/domain/entities"
	"gombot/pkg/handlers"
	"gombot/pkg/repositories"
)

func ExecuteMonitoringOfDoneJobs(ctx context.Context, b *bot.Bot) {
	for {
		job, err := repositories.GetFirstDoneJob()
		if err != nil {
			sleep()
			continue
		}
		// doing some logic
		handlers.SendMessage(b, ctx, job.ChatId, fmt.Sprintf("Job (code:%d) finished successfully", job.ID))
		job.Status = entities.Finished
		repositories.UpdateJob(job)
	}
}
