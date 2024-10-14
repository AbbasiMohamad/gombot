package main

import (
	"context"
	"github.com/go-telegram/bot"
	"gombot/pkg/configs"
	"gombot/pkg/handlers"
	"gombot/pkg/models"
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
		bot.WithCallbackQueryDataHandler("button", bot.MatchTypePrefix, callbackHandler),
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
	// answering callback query first to let Telegram know that we received the callback query,
	// and we're handling it. Otherwise, Telegram might retry sending the update repetitively
	// as it thinks the callback query doesn't reach to our application. learn more by
	// reading the footnote of the https://core.telegram.org/bots/api#callbackquery type.
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text:   "You selected the button: " + update.CallbackQuery.Data,
	})
}

func NoName(ctx context.Context, b *bot.Bot) {
	for {
		job, err := models.PopFromQueue()
		if err != nil {
			log.Printf("Error getting job from queue: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		// decide based on application status

		switch job.Status {
		case models.Requested:
			// make a message to send
			sendMessageForApprove(*job, ctx, b)
			for _, app := range job.Applications {
				app.Status = models.Pending
			}
			job.Status = models.NeedToApproved
			break
		case models.Confirmed:
			// TODO: must to integrate with gitlab
			// validate approvers is staisfied
			// execute logic and update application status
			break
		case models.Done:
			break
		default:
			log.Printf("Unrecognized job status: %v", job.Status)
		}

		time.Sleep(2 * time.Second)

	}
}

// TODO: use pointer?!
func sendMessageForApprove(job models.Job, ctx context.Context, b *bot.Bot) {
	message := makeMessageForApprove(job)

	handlers.SendMessageWithInlineKeyboardMarkup(b, ctx, job.ChatId, message)
}

func makeMessageForApprove(job models.Job) string {
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
