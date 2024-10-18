package handlers

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"gombot/pkg/domain/entities"
	"log"
	"strings"
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

func EditMessageWithInlineKeyboardMarkup(b *bot.Bot, ctx context.Context, chatId int64, message string, messageId int) *models.Message {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "تایید نسخه گذاری", CallbackData: "confirm"},
			},
		},
	}
	result, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      chatId,
		MessageID:   messageId,
		Text:        message,
		ReplyMarkup: kb,
	})
	if err != nil {
		log.Println("There is a unknown error, Gombot can not send message!")
		// TODO: maybe is better to return error to caller
	}
	return result
}

func MakeRequestMessageContent(job *entities.Job) string {
	sb := requestMessageContentHeader(job)
	for _, approver := range job.Approvers {
		if approver.IsApproved {
			sb.WriteString("▫️وضعیت درخواست: *تایید شده*\n")
		} else {
			sb.WriteString("▫️وضعیت درخواست: *در انتظار تایید*\n")
		}
	}
	sb.WriteString("▫️لیست تایید کنندگان مجاز:\n")
	for _, approver := range job.Approvers {
		if approver.IsApproved {
			sb.WriteString(fmt.Sprintf("▪️آقای *%s* (@%s) | وضعیت: *تایید* 🟢\n", approver.FullName, approver.Username))
		} else {
			sb.WriteString(fmt.Sprintf("▪️آقای *%s* (@%s) | وضعیت: *عدم تایید* ⚪\n", approver.FullName, approver.Username))
		}
	}
	sb = requestMessageContentFooter(sb, job)
	return sb.String()
}

func requestMessageContentHeader(job *entities.Job) *strings.Builder {
	sb := strings.Builder{}
	year, month, day := job.CreatedAt.Date()
	sb.WriteString(fmt.Sprintf("✉️ *اعلان درخواست نسخه‌گذاری در تاریخ %v/%v/%v*", year, month, day))
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("بدینوسیله اطلاع داده می‌شود که آقای *%s*(@%s) درخواستِ آپدیت برای ماژول‌(های) زیر را دارد.", job.Requester.FullName, job.Requester.Username))
	sb.WriteString("\n\n")

	for _, application := range job.Applications {
		sb.WriteString(fmt.Sprintf("▪ *%s* (%s) \n", application.PersianName, application.Name))
	}

	sb.WriteString("\n\n")
	return &sb
}

func requestMessageContentFooter(sb *strings.Builder, job *entities.Job) *strings.Builder {
	sb.WriteString("\n")
	sb.WriteString("\n-------------------------------------------\n")
	sb.WriteString("▫️ برای کنسل کردن درخواست دستور زیر را وارد کنید:\n")
	sb.WriteString(fmt.Sprintf("/cancel  %d", job.ID))
	return sb
}

func checkUserHasBaleUsername(userUsername string) bool {
	if userUsername == "" {
		return false
	}
	return true
}
