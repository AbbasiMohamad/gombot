package handlers

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
)

func SendMessage(b *bot.Bot, ctx context.Context, chatId int64, message string) (*models.Message, error) {
	result, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatId,
		Text:   message,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func SendMessageWithInlineKeyboardMarkup(b *bot.Bot, ctx context.Context, chatId int64, message string) *models.Message {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "تایید نسخه گذاری", CallbackData: "confirm"},
			},
		},
	}
	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatId,
		Text:        message,
		ReplyMarkup: kb,
	})
	if err != nil {
		log.Println("There is a unknown error, Gombot can not send message!")
		// TODO: maybe is better to return error to caller
	}
	return msg
}

func EditMessage(b *bot.Bot, ctx context.Context, chatId int64, message string, messageId int) *models.Message {
	result, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    chatId,
		MessageID: messageId,
		Text:      message,
	})
	if err != nil {
		log.Println("There is a unknown error, Gombot can not send message!")
		// TODO: maybe is better to return error to caller
	}
	return result
}

func checkUserHasBaleUsername(userUsername string) bool {
	if userUsername == "" {

		return false
	}
	return true
}
