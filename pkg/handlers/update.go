package handlers

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"gombot/pkg/configs"
	model "gombot/pkg/entities"
	"log"
	"strings"
)

func UpdateHandler(ctx context.Context, b *bot.Bot, update *models.Update) {

	var message string
	// check username existance
	if update.Message.From.Username == "" {
		message = "برای استفاده از بات لازم است نام کاربری داشته باشید"
		SendMessage(b, ctx, update.Message.Chat.ID, message)
		return
	}
	// check access to do update
	isAuthorized := checkAccessToDoUpdate(update.Message.From.Username)
	if !isAuthorized {
		log.Printf("user try to update command failed, because of lack of permission. (user_id:%d|user_fullname:%s)",
			update.Message.From.ID, update.Message.From.FirstName+update.Message.From.LastName)
		message = "شما (" + update.Message.From.FirstName + update.Message.From.LastName + ") دسترسی دستور آپدیت را ندارید"
		SendMessage(b, ctx, update.Message.Chat.ID, message)
		return
	}

	// parse command
	applicationNames := parseCommand(update.Message.Text)
	if len(applicationNames) == 0 {
		message = "اپلیکیشن معتبری برای آپدیت انتخاب نشده است"
		SendMessage(b, ctx, update.Message.Chat.ID, message)
		return
	}
	// validate application existance
	var applications []model.Application
	config, _ := configs.LoadConfig(configs.ConfigPath)
	for _, application := range applicationNames {
		if validateMicroserviceIsExist(application, config.Microservices) {
			app := model.Application{
				Name:   application,
				Status: model.Declared,
			}
			applications = append(applications, app)
		}
	}

	var approvers []model.Approver
	for _, approver := range config.Approvers {
		approvers = append(approvers, model.Approver{
			Username: approver,
		})
	}
	requester := model.Requester{
		Username: update.Message.From.Username,
	}
	job := model.Job{
		JobId:        uuid.New(),
		ChatId:       update.Message.Chat.ID,
		MessageId:    update.Message.ID,
		Applications: applications,
		Status:       model.Requested,
		Approvers:    approvers,
		Requester:    requester,
	}
	// need to be approve
	model.PushToQueue(&job)
}

func UpdateCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
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

// TODO: use pointer?!
func SendMessageForApprove(job model.Job, ctx context.Context, b *bot.Bot) int {
	message := makeMessageForApprove(job)
	msg := SendMessageWithInlineKeyboardMarkup(b, ctx, job.ChatId, message)
	return msg.ID
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
