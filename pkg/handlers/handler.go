package handlers

import (
	"context"
	"gombot/pkg/configs"
	model "gombot/pkg/models"
	"log"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
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

func StatusHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	for _, applications := range model.Queue {
		for _, app := range applications.Applications {
			SendMessage(b, ctx, update.Message.Chat.ID, app.Name)
		}
	}
}

func HelloHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      "Hello, *" + bot.EscapeMarkdown(update.Message.From.FirstName) + "*",
		ParseMode: models.ParseModeMarkdown,
	})
}

func DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Say /hello",
	})
}

func SendMessage(b *bot.Bot, ctx context.Context, chatId int64, message string) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatId,
		Text:   message,
	})
	if err != nil {
		log.Println("There is a unknown error, Gombot can not send message!")
		// TODO: maybe is better to return error to caller
	}
}

func SendMessageWithInlineKeyboardMarkup(b *bot.Bot, ctx context.Context, chatId int64, message string) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "تایید نسخه گذاری", CallbackData: "confirm"},
			},
		},
	}
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatId,
		Text:        message,
		ReplyMarkup: kb,
	})
	if err != nil {
		log.Println("There is a unknown error, Gombot can not send message!")
		// TODO: maybe is better to return error to caller
	}
}

func checkAccessToDoUpdate(requester string) bool {
	config, err := configs.LoadConfig(configs.ConfigPath)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}

	for _, authenticRequester := range config.Requesters {
		if requester == authenticRequester {
			return true
		}
	}
	return false
}

func parseCommand(s string) []string {
	parsedNames := strings.Split(s, "\n")
	var applications []string
	for _, name := range parsedNames[1:] {
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

func validateMicroserviceIsExist(s string, microservices []configs.Microservice) bool {
	for _, item := range microservices {
		if item.Name == s {
			return true
		}
	}
	return false
}
