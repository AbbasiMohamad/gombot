package controllers

import (
	"context"
	"github.com/go-telegram/bot"
	"gombot/pkg/domain/dtos/parameters"
	"gombot/pkg/domain/entities"
	"gombot/pkg/handlers"
	"gombot/pkg/repositories"
	"log"
)

const RETRY_COUNT = 3

func ExecuteMonitoringOfConfirmedJobs(ctx context.Context, b *bot.Bot) {
	for {
		job, err := repositories.GetFirstConfirmedJob()
		if err != nil {
			sleep()
		} else {
			if repositories.IsInProgressJobExists() {
				continue
			}
			handleConfirmedJob(job, ctx, b) // TODO what if app can not start a pipeline?
		}
	}
}

func handleConfirmedJob(job *entities.Job, ctx context.Context, b *bot.Bot) {
	matchApplicationWithGitlabProject(job)
	createPipeline(job)
	messageId := sendStatusMessage(job, ctx, b)
	err := job.SetStatusMessageID(messageId)
	if err != nil {
		log.Printf("can not set status message id: %v", err)
	}
	repositories.UpdateJob(job)
}

func createPipeline(job *entities.Job) {
	retry := RETRY_COUNT
	for retry > 0 {
		for i := range job.Applications {
			pipeline := repositories.CreatePipeline(job.Applications[i].GitlabProjectID, job.Applications[i].Branch)
			if pipeline == nil {
				log.Printf("can not create new pipeline for application '%s'.", job.Applications[i].Name)
				retry--
			} else {
				retry = 0
				p, _ := entities.CreatePipeline(parameters.CreatePipelineParameters{
					PipelineID: pipeline.ID,
					Status:     pipeline.Status,
					Ref:        pipeline.Ref,
					WebURL:     pipeline.WebURL,
				})
				err := job.Applications[i].AddPipeline(&p)
				if err != nil {
					log.Printf("can not create new pipeline for application '%s'.", job.Applications[i].Name)
				}
			}
		}
	}
}

func matchApplicationWithGitlabProject(job *entities.Job) {
	retry := RETRY_COUNT
	for retry > 0 {
		for i := range job.Applications {
			gitlabProject := repositories.GetGitlabProjectByApplicationName(job.Applications[i].Name)
			if gitlabProject == nil {
				log.Printf("there is inconsistency between gitlab projects and application configurations for '%s'.", job.Applications[i].Name)
				job.Applications[i].SetStatusToFailed()
				continue
			}
			job.Applications[i].SetStatusToProcessing()
			job.Applications[i].GitlabProjectID = gitlabProject.ID
		}
		retry--
	}
}

func sendStatusMessage(job *entities.Job, ctx context.Context, b *bot.Bot) int {
	message := handlers.MakeStatusMessageContent(job)
	msg, _ := handlers.SendMessage(b, ctx, job.ChatId, message)
	return msg.ID
}
