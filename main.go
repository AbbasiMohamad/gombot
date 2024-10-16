package main

import (
	"context"
	"github.com/go-telegram/bot"
	"gombot/pkg/configs"
	"gombot/pkg/entities"
	"gombot/pkg/handlers"
	"gombot/pkg/repositories"
	"log"
	"os"
	"os/signal"
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
*/

const (
	SleepTime = 2 * time.Second
)

func main() {
	config := configs.LoadConfig(configs.ConfigPath)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handlers.DefaultHandler),
		bot.WithCallbackQueryDataHandler("confirm", bot.MatchTypeExact, handlers.UpdateCallbackHandler),
	}

	b, err := bot.New(config.Token, opts...)
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
					if jobs[i].MessageId == 0 {
						jobs[i].MessageId = messageId
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
			// TODO: must to integrate with gitlab
			if !repositories.IsInProgressJob() {
				job.Status = entities.InProgress
				handlers.SendMessage(b, ctx, job.ChatId, "فرایند بیلد اپلیکیشن آغاز شد"+"  "+job.JobId.String())
				// validate approvers is staisfied
				// execute logic and update application status
				doSomething(job)
				repositories.UpdateJob(job)
			}
			break
		case entities.Done:
			handlers.SendMessage(b, ctx, job.ChatId, "دیپلوی با موفقیت انجام شد"+"  "+job.JobId.String())
			job.Status = entities.Finished
			// dequeue from queue
			err := entities.DequeueLastItemFromQueue()
			if err != nil {
				log.Println("trying to dequeue from empty queue")
			}
			repositories.UpdateJob(job)
			break
		default:
			log.Printf("Unrecognized job status: %v", job.Status) // TODO: Its better to delete this
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
	time.Sleep(15 * time.Second)
	job.Status = entities.Done
}
func sleep() {
	time.Sleep(SleepTime)
}
