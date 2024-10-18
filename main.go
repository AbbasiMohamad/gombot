package main

import (
	"context"
	"github.com/go-telegram/bot"
	"gombot/pkg/configs"
	"gombot/pkg/controllers"
	"gombot/pkg/handlers"
	"log"
	"os"
	"os/signal"
)

/*
todo and notes
https://ramadhansalmanalfarisi8.medium.com/how-to-dockerize-your-api-with-go-postgresql-gin-docker-9a2b16548520
postgres run command: docker run --name postgres -e POSTGRES_USER=gombot -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres
+ cancel delete the request message
+ audit property for domain
+ improve requester fields
+ study about shadow variables
+ use option pattern for passing message parameters
+ show information of push to request message
+ what if application goes down during building application?
+ set pending status for job that stuck in runner
+ if application crash during InProgress job. it stuck in crash mode
+ create /kill command to handle dangling jobs
+ make SendMessage in handler unreachable for outside of package
*/

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

	go controllers.ExecuteMonitoringOfRequestedJobs(ctx, b)
	go controllers.ExecuteMonitoringOfApprovedJobs(ctx, b)
	go controllers.ExecuteMonitoringOfDoneJobs(ctx, b)
	go controllers.ExecuteMonitoringOfInProgressJobs(ctx, b)

	log.Println("Bot is now running.  Press CTRL-C to exit.")
	b.Start(ctx)
}
