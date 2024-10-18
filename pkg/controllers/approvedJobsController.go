package controllers

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"gombot/pkg/domain/entities"
	"gombot/pkg/handlers"
	"gombot/pkg/repositories"
	"log"
	"strings"
	"time"
)

func ExecuteMonitoringOfApprovedJobs(ctx context.Context, b *bot.Bot) {
	for {
		job, err := repositories.GetFirstApprovedJob()
		if err != nil {
			sleep()
			continue
		}

		// decide based on application status
		switch job.Status {

		case entities.Confirmed:
			if !repositories.IsInProgressJob() {
				job.Status = entities.InProgress
				doSomething(job) // TODO what if app can not start a pipeline?
				// send message to user for monitor pipeline execution
				msg, _ := handlers.SendMessage(b, ctx, job.ChatId, MakeStatusMessageContent(job))
				job.StatusMessageID = msg.ID
				repositories.UpdateJob(job)
			}
			break
		case entities.Done:
			handlers.SendMessage(b, ctx, job.ChatId, "دیپلوی با موفقیت انجام شد"+"  ")
			job.Status = entities.Finished
			if err != nil {
				log.Println("trying to dequeue from empty queue")
			}
			repositories.UpdateJob(job)
			break
		default:
			//log.Printf("Unrecognized job status: %v", job.Status) // TODO: Its better to delete this
		}
		sleep()
	}
}

// TODO: make a meaningful name
func doSomething(job *entities.Job) {
	// match information from gitlab to applications
	for i := range job.Applications {
		job.Applications[i].Status = entities.Processing
		gitlabProject := repositories.GetProjectByApplicationName(job.Applications[i].Name)
		if gitlabProject == nil {
			log.Printf("there is inconsistency between gitlab projects and application configurations for '%s'.", job.Applications[i].Name)
			job.Applications[i].Status = entities.Failed
			continue
		}
		job.Applications[i].GitlabProjectID = gitlabProject.ID
	}
	repositories.UpdateJob(job)
	// iterate on application and create pipeline with 3 retry
	retry := 3
	for retry > 0 {
		for i := range job.Applications {
			pipeline := repositories.CreatePipeline(job.Applications[i].GitlabProjectID, job.Applications[i].Branch)
			if pipeline == nil {
				retry--
			} else {
				retry = 0
				job.Applications[i].Pipeline = entities.Pipeline{
					Status:     pipeline.Status,
					CreatedAt:  time.Now(),
					Ref:        pipeline.Ref,
					PipelineID: pipeline.ID,
					WebURL:     pipeline.WebURL,
				}
				repositories.UpdateJob(job) // save pipeline information
			}

		}
	}

}

func MakeStatusMessageContent(job *entities.Job) string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("*فرایند دیپلوی آغاز شد* (درخواست شماره %d)", job.ID))
	sb.WriteString("\n")
	sb.WriteString("در زیر وضعیت لحظه‌ای دیپلوی اپلیکیشن(ها) گزارش می‌شود.\n")
	sb.WriteString("\n\n")

	for _, application := range job.Applications {
		sb.WriteString(fmt.Sprintf("▪ *%s* (%s): \n", application.PersianName, application.Name))
		var status string
		if application.Pipeline.Status == "created" {
			status = fmt.Sprintf("وضعیت: (%s) 🟡", application.Pipeline.Status)
		} else if application.Pipeline.Status == "running" {
			status = fmt.Sprintf("وضعیت: (%s) 🔵", application.Pipeline.Status)
		} else if application.Pipeline.Status == "success" {
			status = fmt.Sprintf("وضعیت: (%s) 🟢", application.Pipeline.Status)
		} else {
			status = fmt.Sprintf("وضعیت: (%s) 🔴", application.Pipeline.Status)
		}
		sb.WriteString(status)
		if application.Pipeline.Status != "success" || application.Pipeline.Status != "failed" || application.Pipeline.Status != "canceled" || application.Pipeline.Status != "skipped" {
			sb.WriteString(fmt.Sprintf(" \t%s\t", calculateDurationTime(application.Pipeline.CreatedAt)))
		} else {
			sb.WriteString(fmt.Sprintf("\n مدت زمان طول کشیده: %s", calculatePipelineDuration(application.Pipeline)))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n\n")
	sb.WriteString("\n")

	return sb.String()
}

func calculatePipelineDuration(pipeline entities.Pipeline) string { //TODO: duplicate code
	// Get the current time
	now := pipeline.FinishedAt

	// Calculate the duration between the current time and the given time `t`
	duration := now.Sub(pipeline.CreatedAt)

	// Determine the time unit to return (seconds, minutes, hours, days, weeks)
	switch {
	case duration < time.Minute:
		// Less than a minute ago
		return fmt.Sprintf("%d ثانیه", int(duration.Seconds()))
	case duration < time.Hour:
		// Less than an hour ago
		return fmt.Sprintf("%d دقیقه", int(duration.Minutes()))
	case duration < 24*time.Hour:
		// Less than a day ago
		return fmt.Sprintf("%d ساعت", int(duration.Hours()))
	case duration < 7*24*time.Hour:
		// Less than a week ago
		return fmt.Sprintf("%d روز", int(duration.Hours()/24))
	case duration < 30*24*time.Hour:
		// Less than a month ago
		return fmt.Sprintf("%d هفته", int(duration.Hours()/(24*7)))
	default:
		// More than a month ago
		months := int(duration.Hours() / (24 * 30))
		return fmt.Sprintf("%d ماه", months)
	}
}

func calculateDurationTime(t time.Time) string {
	// Get the current time
	now := time.Now()

	// Calculate the duration between the current time and the given time `t`
	duration := now.Sub(t)

	// Determine the time unit to return (seconds, minutes, hours, days, weeks)
	switch {
	case duration < time.Minute:
		// Less than a minute ago
		return fmt.Sprintf("%d ثانیه قبل", int(duration.Seconds()))
	case duration < time.Hour:
		// Less than an hour ago
		return fmt.Sprintf("%d دقیقه قبل", int(duration.Minutes()))
	case duration < 24*time.Hour:
		// Less than a day ago
		return fmt.Sprintf("%d ساعت قبل", int(duration.Hours()))
	case duration < 7*24*time.Hour:
		// Less than a week ago
		return fmt.Sprintf("%d روز قبل", int(duration.Hours()/24))
	case duration < 30*24*time.Hour:
		// Less than a month ago
		return fmt.Sprintf("%d هفته قبل", int(duration.Hours()/(24*7)))
	default:
		// More than a month ago
		months := int(duration.Hours() / (24 * 30))
		return fmt.Sprintf("%d months ago", months)
	}
}
func calculateDurationTimeEN(t time.Time) string {
	// Get the current time
	now := time.Now()

	// Calculate the duration between the current time and the given time `t`
	duration := now.Sub(t)

	// Determine the time unit to return (seconds, minutes, hours, days, weeks)
	switch {
	case duration < time.Minute:
		// Less than a minute ago
		return fmt.Sprintf("%d seconds ago", int(duration.Seconds()))
	case duration < time.Hour:
		// Less than an hour ago
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	case duration < 24*time.Hour:
		// Less than a day ago
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	case duration < 7*24*time.Hour:
		// Less than a week ago
		return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	case duration < 30*24*time.Hour:
		// Less than a month ago
		return fmt.Sprintf("%d weeks ago", int(duration.Hours()/(24*7)))
	default:
		// More than a month ago
		months := int(duration.Hours() / (24 * 30))
		return fmt.Sprintf("%d months ago", months)
	}
}
