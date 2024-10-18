package controllers

import (
	"context"
	"github.com/go-telegram/bot"
	"gombot/pkg/domain/entities"
	"gombot/pkg/handlers"
	"gombot/pkg/repositories"
	"log"
	"time"
)

func ExecuteMonitoringOfInProgressJobs(ctx context.Context, b *bot.Bot) {
	for {
		job, err := repositories.GetFirstInProgressJob()
		if err != nil {
			sleep()
			continue
		}

		switch job.Status {
		case entities.InProgress:
			// check pipelines status and update message
			sleep()
			updatePipelinesStatus(job)
			handlers.EditMessage(b, ctx, job.ChatId, MakeStatusMessageContent(job), job.StatusMessageID)
			updateJobStatus(job)
			sleep()
		}
	}

}

func updatePipelinesStatus(job *entities.Job) {
	for i, _ := range job.Applications {
		pipeline := repositories.GetPipeline(job.Applications[i].GitlabProjectID, job.Applications[i].Pipeline.PipelineID)
		if pipeline == nil {
			log.Printf("there is a error to get pipeline for '%s' application", job.Applications[i].Name)
			continue
		}
		job.Applications[i].Pipeline.Status = pipeline.Status
		switch pipeline.Status {
		case "success":
			job.Applications[i].Status = entities.Deployed
			job.Applications[i].Pipeline.FinishedAt = time.Now()
			break
		case "failed":
		case "canceled":
		case "skipped":
			job.Applications[i].Status = entities.Failed
			job.Applications[i].Pipeline.FinishedAt = time.Now()
			break
		default:
			job.Applications[i].Status = entities.Processing
		}
		repositories.NewUpdateJob(job)
	}
}

func updateJobStatus(job *entities.Job) {
	isJobFinished := true
	for _, app := range job.Applications {
		switch app.Status {
		case entities.Processing:
			isJobFinished = false
			break
		}
	}
	if isJobFinished {
		job.Status = entities.Done
		repositories.NewUpdateJob(job)
	}
}
