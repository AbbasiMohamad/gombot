package main

import (
	"context"
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

	b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, handlers.HelloHandler)
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
		message = "You are authorized to perform this action."
	} else {
		message = "You are not authorized to perform this action."
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text:   message,
	})
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

func NoName(ctx context.Context, b *bot.Bot) {
	for {
		job, err := model.PopFromQueue()
		if err != nil {
			log.Printf("Error getting job from queue: %v", err)
			time.Sleep(2 * time.Second)
			continue
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
			// validate approvers is staisfied
			// execute logic and update application status
			break
		case model.Done:
			break
		default:
			log.Printf("Unrecognized job status: %v", job.Status)
		}

		time.Sleep(2 * time.Second)

	}
}

// TODO: use pointer?!
func sendMessageForApprove(job model.Job, ctx context.Context, b *bot.Bot) {
	message := makeMessageForApprove(job)
	handlers.SendMessageWithInlineKeyboardMarkup(b, ctx, job.ChatId, message)
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
	sb.WriteString("تایید کنندگان: \n")
	for _, approver := range job.Approvers {
		sb.WriteString("@" + approver.Username + "\n")
	}
	sb.WriteString("\n-----------------------\n")
	return sb.String()
}
