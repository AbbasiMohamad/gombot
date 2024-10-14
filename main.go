package main

import (
	"context"
	"github.com/go-telegram/bot"
	"gombot/pkg/configs"
	"gombot/pkg/handlers"
	"log"
	"os"
	"os/signal"
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
	}

	b, err := bot.New(config.Token, opts...)
	if nil != err {
		// TODO: handle error properly
		panic(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/hello", bot.MatchTypeExact, handlers.HelloHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/update", bot.MatchTypePrefix, handlers.UpdateHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/status", bot.MatchTypePrefix, handlers.StatusHandler)

	b.Start(ctx)
}
