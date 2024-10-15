package main

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"gombot/pkg/configs"
	"gombot/pkg/handlers"
	model "gombot/pkg/models"
	"log"
	"os"
	"os/signal"
	"strings"
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
		bot.WithCallbackQueryDataHandler("confirm", bot.MatchTypeExact, callbackHandler),
	}

	b, err := bot.New(config.Token, opts...)
	if nil != err {
		// TODO: handle error properly
		panic(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, handlers.HelpHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/update", bot.MatchTypePrefix, handlers.UpdateHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/status", bot.MatchTypePrefix, handlers.StatusHandler)

	go NoName(ctx, b)
	log.Println("Bot is now running.  Press CTRL-C to exit.")
	b.Start(ctx)
}

func callbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	var message string
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       true,
	})
	isAuthorized := checkAccessToDoApprove(update.CallbackQuery.From.Username)
	if isAuthorized {
		job, _ := model.PopJobByMessageIdFromQueue(update.CallbackQuery.Message.Message.ID) //TODO: error handling! and change method
		if job.ChatId == update.CallbackQuery.Message.Message.Chat.ID {
			for i, _ := range job.Approvers {
				if job.Approvers[i].Username == update.CallbackQuery.From.Username {
					job.Approvers[i].Approved = true
				}
			}
			message = makeMessageForApprove(*job)
			allApproved := true
			for _, approver := range job.Approvers {
				if !approver.Approved {
					allApproved = false
				}
			}

			if allApproved {
				job.Status = model.Confirmed
				b.EditMessageText(ctx, &bot.EditMessageTextParams{
					ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
					MessageID: update.CallbackQuery.Message.Message.ID,
					Text:      message,
				})
			} else {
				kb := &models.InlineKeyboardMarkup{
					InlineKeyboard: [][]models.InlineKeyboardButton{
						{
							{Text: "تایید نسخه گذاری", CallbackData: "confirm"},
						},
					},
				}
				b.EditMessageText(ctx, &bot.EditMessageTextParams{
					ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
					MessageID:   update.CallbackQuery.Message.Message.ID,
					Text:        message,
					ReplyMarkup: kb,
				})
			}

		}

	} else {
		message = "You are not authorized to perform this action."
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.CallbackQuery.Message.Message.Chat.ID,
			Text:   message,
		})
	}

}

func checkAccessToDoApprove(approver string) bool {
	config, err := configs.LoadConfig(configs.ConfigPath)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}

	for _, authenticApprover := range config.Approvers {
		if approver == authenticApprover {
			return true
		}
	}
	return false
}
func sleep() {
	time.Sleep(SleepTime)
}
func NoName(ctx context.Context, b *bot.Bot) {
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
					messageId := sendMessageForApprove(*jobs[i], ctx, b)
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
			sendMessageForApprove(*job, ctx, b)
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

// TODO: use pointer?!
func sendMessageForApprove(job model.Job, ctx context.Context, b *bot.Bot) int {
	message := makeMessageForApprove(job)
	msg := handlers.SendMessageWithInlineKeyboardMarkup(b, ctx, job.ChatId, message)
	return msg.ID
}

func makeMessageForApprove(job model.Job) string {
	sb := strings.Builder{}
	sb.WriteString("\nدرخواست نسخه گذاری برای اپلیکیشن های زیر:\n")
	for _, application := range job.Applications {
		sb.WriteString(application.Name + "\n")
	}
	sb.WriteString("\n-----------------------\n")
	sb.WriteString("کد درخواست: " + job.JobId.String())
	sb.WriteString("\n-----------------------\n")
	sb.WriteString("درخواست کننده: " + job.Requester.Username)
	sb.WriteString("\n-----------------------\n")
	sb.WriteString("وضعیت درخواست: در انتظار تایید کنندگان ")
	sb.WriteString("\n-----------------------\n")
	sb.WriteString("تایید کنندگان: \n" + fmt.Sprintf("\n(%d/%d)\n", 0, len(job.Approvers)))
	for _, approver := range job.Approvers {
		if approver.Approved {
			sb.WriteString("@" + approver.Username + " 👍\n")
		} else {
			sb.WriteString("@" + approver.Username + "\n")
		}

	}
	sb.WriteString("\n-----------------------\n")
	return sb.String()
}
