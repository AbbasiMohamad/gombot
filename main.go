package main

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"gombot/pkg/configs"
	"gombot/pkg/entities"
	"gombot/pkg/handlers"
	"gombot/pkg/repositories"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
)

/*
todo and notes
https://ramadhansalmanalfarisi8.medium.com/how-to-dockerize-your-api-with-go-postgresql-gin-docker-9a2b16548520
postgres run command: docker run --name postgres -e POSTGRES_USER=gombot -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres
+ cancel delete the request message
+ audit property for entities
+ improve requester fields
+ study about shadow variables
+ use option pattern for passing message parameters
+ show information of push to request message
+ what if application goes down during building application?
+ set pending status for job that stuck in runner
+ if application crash during InProgress job. it stuck in crash mode
*/

const (
	SleepTime = 2 * time.Second
)

func main() {
	config := configs.LoadConfig(configs.ConfigPath)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithServerURL("https://tapi.bale.ai"),
		bot.WithDefaultHandler(handlers.DefaultHandler),
		bot.WithCallbackQueryDataHandler("confirm", bot.MatchTypeExact, handlers.UpdateCallbackHandler),
	}

	b, err := bot.New(config.BaleToken, opts...)
	if nil != err {
		// TODO: handle error properly
		panic(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, handlers.HelpHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/update", bot.MatchTypePrefix, handlers.UpdateHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/status", bot.MatchTypePrefix, handlers.StatusHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/test", bot.MatchTypePrefix, handlers.TestHandler)

	go executeMonitoringOfRequestedJobs(ctx, b)
	go executeMonitoringOfApprovedJobs(ctx, b)
	go executeMonitoringOfDoneJobs(ctx, b)

	log.Println("Bot is now running.  Press CTRL-C to exit.")
	b.Start(ctx)
}

func executeMonitoringOfRequestedJobs(ctx context.Context, b *bot.Bot) {
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

func executeMonitoringOfApprovedJobs(ctx context.Context, b *bot.Bot) {
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
				doSomething(job) // TODO what if app dose not start a pipeline?
				msg := handlers.SendMessage(b, ctx, job.ChatId, makeStatusMessageContent(job))
				job.StatusMessageID = msg.ID
				repositories.UpdateJob(job)
			}
			break
		case entities.Done:
			handlers.SendMessage(b, ctx, job.ChatId, "دیپلوی با موفقیت انجام شد"+"  "+job.JobId.String())
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

func executeMonitoringOfDoneJobs(ctx context.Context, b *bot.Bot) {
	for {
		job, err := repositories.GetFirstDoneJob()
		if err != nil {
			sleep()
			continue
		}
		// doing some logic
		handlers.SendMessage(b, ctx, job.ChatId, "دیپلوی با موفقیت انجام شد"+"  "+job.JobId.String())
		job.Status = entities.Finished
		repositories.UpdateJob(job)
	}
}

func doSomething(job *entities.Job) {
	// match information from gitlab to applications
	for i, _ := range job.Applications {
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
		for i, _ := range job.Applications {
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
				repositories.UpdateJob(job)
			}

		}
	}

	// save pipeline information

	// send message to user for monitor pipeline execution

	// create another controller to manage status of jobs

	time.Sleep(15 * time.Second)
	job.Status = entities.Done

}
func sleep() {
	time.Sleep(SleepTime)
}

/*
💠 *فرایند دیپلوی اپلیکیشن آغاز شد* (درخواست شماره 9)
در زیر وضعیت لحظه‌ای دیپلوی اپلیکیشن(ها) گزارش می‌شود.

▪️ اطلاعات پایه (mdm) | وضعیت: *در حال دیپلوی* 🔵
▪️ گمبات (my-project) | وضعیت: *دیپلوی شده* 🟢
*/
func makeStatusMessageContent(job *entities.Job) string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("*فرایند دیپلوی اپلیکیشن آغاز شد* (درخواست شماره %d)", job.ID))
	sb.WriteString("\n")
	sb.WriteString("در زیر وضعیت لحظه‌ای دیپلوی اپلیکیشن(ها) گزارش می‌شود.\n")
	sb.WriteString("\n\n")

	for _, application := range job.Applications {
		sb.WriteString(fmt.Sprintf("▪ *%s* (%s): \n", application.PersianName, application.Name))
		sb.WriteString(fmt.Sprintf("وضعیت: درحال دیپلوی 🔵\n"))
		sb.WriteString(fmt.Sprintf("مدت زمان سپری شده: 6 دقیقه\n"))
		sb.WriteString("\n")
	}
	sb.WriteString("\n\n")
	sb.WriteString("\n")

	return sb.String()
}
