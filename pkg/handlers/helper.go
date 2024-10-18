package handlers

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"gombot/pkg/domain/entities"
	"log"
	"strings"
	"time"
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

func MakeStatusMessageContent(job *entities.Job) string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("*فرایند دیپلوی آغاز شد* (درخواست شماره %d)", job.ID))
	sb.WriteString("\n")
	sb.WriteString("در زیر وضعیت لحظه‌ای دیپلوی اپلیکیشن(ها) گزارش می‌شود.\n")
	sb.WriteString("\n\n")

	for _, application := range job.Applications {
		sb.WriteString(fmt.Sprintf("▪ *%s* (%s): \n", application.PersianName, application.Name))
		var status string
		if application.Pipeline.Status == "created" {
			status = fmt.Sprintf("وضعیت: (%s) 🟡", application.Pipeline.Status)
		} else if application.Pipeline.Status == "running" {
			status = fmt.Sprintf("وضعیت: (%s) 🔵", application.Pipeline.Status)
		} else if application.Pipeline.Status == "success" {
			status = fmt.Sprintf("وضعیت: (%s) 🟢", application.Pipeline.Status)
		} else if application.Pipeline.Status == "" {
			status = fmt.Sprintf("وضعیت: (%s) 🟣", application.Pipeline.Status)
		} else {
			status = fmt.Sprintf("وضعیت: (%s) 🔴", application.Pipeline.Status)
		}
		sb.WriteString(status)
		if application.Pipeline.Status != "success" || application.Pipeline.Status != "failed" || application.Pipeline.Status != "canceled" || application.Pipeline.Status != "skipped" {
			sb.WriteString(fmt.Sprintf(" \t%s\t", calculateDurationTime(application.Pipeline.CreatedAt)))
		} else {
			sb.WriteString(fmt.Sprintf("\n مدت زمان طول کشیده: %s", calculatePipelineDuration(application.Pipeline)))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n\n")
	sb.WriteString("\n")

	return sb.String()
}

func calculatePipelineDuration(pipeline *entities.Pipeline) string { //TODO: duplicate code
	// Get the current time
	now := pipeline.FinishedAt

	// Calculate the duration between the current time and the given time `t`
	duration := now.Sub(pipeline.CreatedAt)

	// Determine the time unit to return (seconds, minutes, hours, days, weeks)
	switch {
	case duration < time.Minute:
		// Less than a minute ago
		return fmt.Sprintf("%d ثانیه", int(duration.Seconds()))
	case duration < time.Hour:
		// Less than an hour ago
		return fmt.Sprintf("%d دقیقه", int(duration.Minutes()))
	case duration < 24*time.Hour:
		// Less than a day ago
		return fmt.Sprintf("%d ساعت", int(duration.Hours()))
	case duration < 7*24*time.Hour:
		// Less than a week ago
		return fmt.Sprintf("%d روز", int(duration.Hours()/24))
	case duration < 30*24*time.Hour:
		// Less than a month ago
		return fmt.Sprintf("%d هفته", int(duration.Hours()/(24*7)))
	default:
		// More than a month ago
		months := int(duration.Hours() / (24 * 30))
		return fmt.Sprintf("%d ماه", months)
	}
}

func calculateDurationTime(t time.Time) string {
	// Get the current time
	now := time.Now()

	// Calculate the duration between the current time and the given time `t`
	duration := now.Sub(t)

	// Determine the time unit to return (seconds, minutes, hours, days, weeks)
	switch {
	case duration < time.Minute:
		// Less than a minute ago
		return fmt.Sprintf("%d ثانیه قبل", int(duration.Seconds()))
	case duration < time.Hour:
		// Less than an hour ago
		return fmt.Sprintf("%d دقیقه قبل", int(duration.Minutes()))
	case duration < 24*time.Hour:
		// Less than a day ago
		return fmt.Sprintf("%d ساعت قبل", int(duration.Hours()))
	case duration < 7*24*time.Hour:
		// Less than a week ago
		return fmt.Sprintf("%d روز قبل", int(duration.Hours()/24))
	case duration < 30*24*time.Hour:
		// Less than a month ago
		return fmt.Sprintf("%d هفته قبل", int(duration.Hours()/(24*7)))
	default:
		// More than a month ago
		months := int(duration.Hours() / (24 * 30))
		return fmt.Sprintf("%d months ago", months)
	}
}
func calculateDurationTimeEN(t time.Time) string {
	// Get the current time
	now := time.Now()

	// Calculate the duration between the current time and the given time `t`
	duration := now.Sub(t)

	// Determine the time unit to return (seconds, minutes, hours, days, weeks)
	switch {
	case duration < time.Minute:
		// Less than a minute ago
		return fmt.Sprintf("%d seconds ago", int(duration.Seconds()))
	case duration < time.Hour:
		// Less than an hour ago
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	case duration < 24*time.Hour:
		// Less than a day ago
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	case duration < 7*24*time.Hour:
		// Less than a week ago
		return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	case duration < 30*24*time.Hour:
		// Less than a month ago
		return fmt.Sprintf("%d weeks ago", int(duration.Hours()/(24*7)))
	default:
		// More than a month ago
		months := int(duration.Hours() / (24 * 30))
		return fmt.Sprintf("%d months ago", months)
	}
}
