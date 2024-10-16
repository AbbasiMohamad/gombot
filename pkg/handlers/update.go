package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"gombot/pkg/configs"
	"gombot/pkg/entities"
	"gombot/pkg/repositories"
	"log"
	"strings"
	"time"
)

func UpdateHandler(ctx context.Context, b *bot.Bot, update *models.Update) {

	var message string
	// check username existance
	if update.Message.From.Username == "" {
		message = "برای استفاده از بات لازم است نام کاربری داشته باشید"
		SendMessage(b, ctx, update.Message.Chat.ID, message)
		return
	}

	isAuthorized := checkAccessToDoUpdate(update.Message.From.Username)
	if !isAuthorized {
		log.Printf("user try to update command failed, because of lack of permission. (user_id:%d|user_fullname:%s)",
			update.Message.From.ID, update.Message.From.FirstName+update.Message.From.LastName) // TODO: is it required?
		message = fmt.Sprintf("شما (%s %s) دسترسی درخواست نخسه گذاری ندارید", update.Message.From.FirstName, update.Message.From.LastName)
		SendMessage(b, ctx, update.Message.Chat.ID, message)
		return
	}

	// parse command
	applicationNames := parseUpdateCommand(update.Message.Text)

	// validate application existance
	var applications []entities.Application
	config := configs.LoadConfig(configs.ConfigPath)

	if len(applicationNames) > 0 {
		for _, application := range applicationNames {
			application = strings.ToLower(application)
			if validateMicroserviceIsExist(application, config.Microservices) {
				appConfig, err := getMicroserviceConfig(application, config.Microservices)
				if err != nil {
					log.Printf("there is dangrouse error") // TODO: fix that
				}
				app := entities.Application{
					Name:          application,
					PersianName:   appConfig.PersianName,
					NeedToApprove: appConfig.NeedToApprove,
					Status:        entities.Declared,
				}
				applications = append(applications, app)
			}
		}
	}

	if len(applications) == 0 || applications == nil {
		message = "اپلیکیشن معتبری برای آپدیت انتخاب نشده است"
		SendMessage(b, ctx, update.Message.Chat.ID, message)
		return
	}

	var approvers []entities.Approver
	for _, approver := range config.Approvers {
		approvers = append(approvers, entities.Approver{
			Username: approver.Username,
			FullName: approver.FullName,
		})
	}
	requester := entities.Requester{
		Username: update.Message.From.Username,
		FullName: update.Message.From.FirstName + " " + update.Message.From.LastName,
	}
	job := entities.Job{
		JobId:        uuid.New(),
		ChatId:       update.Message.Chat.ID,
		Applications: applications,
		Status:       entities.Requested,
		Approvers:    approvers,
		Requester:    requester,
		CreatedAt:    time.Now(),
	}
	// need to be approve
	if repositories.InsertJob(&job) > 0 {
		entities.PushToQueue(&job)
	}

}

func UpdateCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	var message string
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       true,
	})
	isAuthorized := checkAccessToDoApprove(update.CallbackQuery.From.Username)
	if isAuthorized {
		job := repositories.GetJobByMessageId(update.CallbackQuery.Message.Message.ID)
		if job.ChatId == update.CallbackQuery.Message.Message.Chat.ID {
			for i, _ := range job.Approvers {
				if job.Approvers[i].Username == update.CallbackQuery.From.Username {
					job.Approvers[i].IsApproved = true
				}
			}
			message = makeMessageForApprove(*job)
			allApproved := true
			for _, approver := range job.Approvers {
				if !approver.IsApproved {
					allApproved = false
				}
			}

			if allApproved {
				job.Status = entities.Confirmed
				b.EditMessageText(ctx, &bot.EditMessageTextParams{
					ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
					MessageID: update.CallbackQuery.Message.Message.ID,
					Text:      message,
				})
				//repositories.UpdateJob(job)
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
				//repositories.UpdateJob(job)
			}

			repositories.UpdateJob(job)

		}

	} else {
		message = "You are not authorized to perform this action."
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.CallbackQuery.Message.Message.Chat.ID,
			Text:   message,
		})
	}

}

// TODO: use pointer?!
func SendMessageForApprove(job entities.Job, ctx context.Context, b *bot.Bot) int {
	message := makeMessageForApprove(job)
	msg := SendMessageWithInlineKeyboardMarkup(b, ctx, job.ChatId, message)
	return msg.ID
}

func checkAccessToDoApprove(approver string) bool {
	config := configs.LoadConfig(configs.ConfigPath)

	for _, authenticApprover := range config.Approvers {
		if approver == authenticApprover.Username {
			return true
		}
	}
	return false
}

func parseUpdateCommand(s string) []string {
	parsedNames := strings.Split(s, "\n")
	var applications []string
	for _, name := range parsedNames {
		parsedApplicationNames := strings.Split(strings.TrimSpace(name), " ")
		if len(parsedApplicationNames) > 1 {
			for _, appName := range parsedApplicationNames {
				applications = append(applications, appName)
			}
		} else if name != "" {
			applications = append(applications, name)
		}
	}
	return removeDuplicates(applications)
}

func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool) // Map to track seen values
	result := []string{}

	for _, val := range slice {
		if _, ok := seen[val]; !ok { // If the value hasn't been seen yet
			seen[val] = true // Mark it as seen
			result = append(result, val)
		}
	}

	return result
}

func checkAccessToDoUpdate(requester string) bool {
	config := configs.LoadConfig(configs.ConfigPath)
	for _, authenticRequester := range config.Requesters {
		if requester == authenticRequester {
			return true
		}
	}
	return false
}

func validateMicroserviceIsExist(s string, microservices []configs.Microservice) bool {
	for _, item := range microservices {
		if item.Name == s {
			return true
		}
	}
	return false
}
func getMicroserviceConfig(s string, microservices []configs.Microservice) (configs.Microservice, error) {
	for _, item := range microservices {
		if item.Name == s {
			return item, nil
		}
	}
	return configs.Microservice{}, errors.New("there is no any corresponding microservice config")
}
func makeMessageForApprove(job entities.Job) string {
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
	sb.WriteString("▫️وضعیت درخواست: *در انتظار تایید*\n")
	sb.WriteString("▫️لیست تایید کنندگان مجاز:\n")

	for _, approver := range job.Approvers {
		if approver.IsApproved {
			sb.WriteString(fmt.Sprintf("▪️آقای *%s* (@%s) | وضعیت: *تایید* 🟢\n", approver.FullName, approver.Username))
		} else {
			sb.WriteString(fmt.Sprintf("▪️آقای *%s* (@%s) | وضعیت: *عدم تایید* ⚪\n", approver.FullName, approver.Username))
		}
	}
	sb.WriteString("\n")
	sb.WriteString("\n---------------------------------------------------------------------\n")
	sb.WriteString("▫️ برای کنسل کردن درخواست دستور زیر را وارد کنید:\n")
	sb.WriteString(fmt.Sprintf("/cancel  %s", job.JobId))
	return sb.String()
}
