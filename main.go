package main

import (
	"context"
	"github.com/go-telegram/bot"
	"gombot/pkg/configs"
	"gombot/pkg/entities"
	"gombot/pkg/handlers"
	"log"
	"os"
	"os/signal"
	"time"
)

/*
TODO and links
https://ramadhansalmanalfarisi8.medium.com/how-to-dockerize-your-api-with-go-postgresql-gin-docker-9a2b16548520
postgres run command: docker run --name postgres -e POSTGRES_USER=gombot -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres
+ cancel delete the request message
+ audit property for entities
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

	go executeMonitoringOfQueue(ctx, b)
	log.Println("Bot is now running.  Press CTRL-C to exit.")
	b.Start(ctx)
}

func executeMonitoringOfQueue(ctx context.Context, b *bot.Bot) {
	for {
		var job *entities.Job
		jobs, err := entities.PopRequestedJobsFromQueue()
		if err != nil {
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
				}
			}
			continue
		} else {
			job, err = entities.PopLastItemFromQueue()
			if err != nil {
				sleep()
				continue
			}
		}

		// decide based on application status
		switch job.Status {
		case entities.Requested:
			// make a message to send
			handlers.SendMessageForApprove(*job, ctx, b)
			for _, app := range job.Applications {
				app.Status = entities.Pending
			}
			job.Status = entities.NeedToApproved
			break
		case entities.Confirmed:
			// TODO: must to integrate with gitlab
			handlers.SendMessage(b, ctx, job.ChatId, "فرایند بیلد اپلیکیشن آغاز شد"+"  "+job.JobId.String())
			job.Status = entities.Done
			// validate approvers is staisfied
			// execute logic and update application status
			break
		case entities.Done:
			handlers.SendMessage(b, ctx, job.ChatId, "دیپلوی با موفقیت انجام شد"+"  "+job.JobId.String())
			job.Status = entities.None
			// dequeue from queue
			err := entities.DequeueLastItemFromQueue()
			if err != nil {
				log.Println("Trying to dequeue from empty queue")
			}
			break
		default:
			log.Printf("Unrecognized job status: %v", job.Status) // TODO: Its better to delete this
		}
		sleep()
	}
}

func sleep() {
	time.Sleep(SleepTime)
}
