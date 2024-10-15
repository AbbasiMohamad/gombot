package main

import (
	"context"
	"github.com/go-telegram/bot"
	"gombot/pkg/configs"
	model "gombot/pkg/entities"
	"gombot/pkg/handlers"
	"log"
	"os"
	"os/signal"
	"time"
)

const (
	SleepTime = 2 * time.Second
)

func main() {
	config, err := configs.LoadConfig(configs.ConfigPath)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}

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

	go executeMonitoringOfQueue(ctx, b)
	log.Println("Bot is now running.  Press CTRL-C to exit.")
	b.Start(ctx)
}

func executeMonitoringOfQueue(ctx context.Context, b *bot.Bot) {
	for {
		var job *model.Job
		jobs, err := model.PopRequestedJobsFromQueue()
		if err != nil {
			sleep() //TODO: make sleep method and read from config
			continue
		}

		if len(jobs) > 0 {
			for i, _ := range jobs {
				if jobs[i].Status == model.Requested {
					messageId := handlers.SendMessageForApprove(*jobs[i], ctx, b)
					for j, _ := range jobs[i].Applications {
						jobs[i].Applications[j].Status = model.Pending
					}
					jobs[i].Status = model.NeedToApproved
					//important: update message id of job
					if jobs[i].MessageId == 0 {
						jobs[i].MessageId = messageId
					}
				}
			}
			continue
		} else {
			job, err = model.PopLastItemFromQueue()
			if err != nil {
				sleep()
				continue
			}
		}

		// decide based on application status
		switch job.Status {
		case model.Requested:
			// make a message to send
			handlers.SendMessageForApprove(*job, ctx, b)
			for _, app := range job.Applications {
				app.Status = model.Pending
			}
			job.Status = model.NeedToApproved
			break
		case model.Confirmed:
			// TODO: must to integrate with gitlab
			handlers.SendMessage(b, ctx, job.ChatId, "فرایند بیلد اپلیکیشن آغاز شد"+"  "+job.JobId.String())
			job.Status = model.Done
			// validate approvers is staisfied
			// execute logic and update application status
			break
		case model.Done:
			handlers.SendMessage(b, ctx, job.ChatId, "دیپلوی با موفقیت انجام شد"+"  "+job.JobId.String())
			job.Status = model.None
			// dequeue from queue
			err := model.DequeueLastItemFromQueue()
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
