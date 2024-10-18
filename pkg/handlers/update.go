package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"gombot/pkg/configs"
	"gombot/pkg/domain/dtos/parameters"
	"gombot/pkg/domain/entities"
	"gombot/pkg/repositories"
	"log"
	"strings"
)

// UpdateHandler checks corresponding permissions and validations and finally creates Job object with status 'requested' and
// Application's status 'pending'
func UpdateHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	var message string

	hasUsername := checkUserHasBaleUsername(update.Message.From.Username)
	if !hasUsername {
		message = "*بمنظور درخواست آپدیت باید اکانت بله دارای نام کاربری باشد.*"
		if _, err := SendMessage(b, ctx, update.Message.Chat.ID, message); err != nil {
			log.Printf("failed to send message: %v", err)
		}
		return
	}

	isAuthorized := checkAccessToDoUpdate(update.Message.From.Username)
	if !isAuthorized {
		message = fmt.Sprintf("شما (%s %s) دسترسی درخواست نخسه گذاری ندارید", update.Message.From.FirstName, update.Message.From.LastName)
		if _, err := SendMessage(b, ctx, update.Message.Chat.ID, message); err != nil {
			log.Printf("failed to send message: %v", err)
		}
		return
	}

	userInputApplicationNames := parseUpdateCommand(update.Message.Text)

	applications := createApplicationsBasedOnUserInputAndConfigurations(userInputApplicationNames)

	if len(applications) == 0 || applications == nil {
		message = "اپلیکیشن معتبری برای آپدیت انتخاب نشده است"
		if _, err := SendMessage(b, ctx, update.Message.Chat.ID, message); err != nil {
			log.Printf("failed to send message: %v", err)
		}
		return
	}

	job, err := createJobWithCorrespondingApproversAndRequester(update, applications)
	if err != nil {
		if _, sendError := SendMessage(b, ctx, update.Message.Chat.ID, err.Error()); sendError != nil {
			log.Printf("failed to send message: %v", err)
		}
		log.Printf("failed to create job: %v", err)
	}

	repositories.InsertJob(&job)
}

// UpdateCallbackHandler checks corresponding permissions and finally make status entities.Confirmed.
func UpdateCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	isAuthorized := checkAccessToDoApprove(update.CallbackQuery.From.Username)
	if !isAuthorized {
		return
	}

	job := repositories.GetJobByMessageId(update.CallbackQuery.Message.Message.ID)
	if job.ChatId != update.CallbackQuery.Message.Message.Chat.ID {
		log.Printf("there is a big problem in GetJobByMessageId query.")
		return
	}

	err := job.SetApproverResponse(update.CallbackQuery.From.Username)
	if err != nil {
		log.Printf("failed to set approver response: %v", err)
	}

	message := MakeRequestMessageContent(job)

	if job.Status == entities.Confirmed {
		EditMessage(b, ctx, update.CallbackQuery.Message.Message.Chat.ID, message, update.CallbackQuery.Message.Message.ID)
	} else {
		EditMessageWithInlineKeyboardMarkup(b, ctx, update.CallbackQuery.Message.Message.Chat.ID, message, update.CallbackQuery.Message.Message.ID)
	}
	repositories.UpdateJob(job)
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

// parseUpdateCommand parse user input as /update command
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
	return removeDuplicates(applications[1:])
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

func getMicroserviceConfig(s string, microservices []configs.Application) (configs.Application, error) {
	for _, item := range microservices {
		if item.Name == s {
			return item, nil
		}
	}
	return configs.Application{}, errors.New("there is no any corresponding microservice config")
}

// create application entities with status declared based on microservice configuration
func createApplicationsBasedOnUserInputAndConfigurations(userInputApplicationNames []string) []entities.Application {
	var applications []entities.Application
	config := configs.LoadConfig(configs.ConfigPath)
	if len(userInputApplicationNames) > 0 {
		for _, application := range userInputApplicationNames {
			application = strings.ToLower(application)
			appConfig, err := getMicroserviceConfig(application, config.Applications)
			if err != nil {
				log.Printf("there is no any corresponding microservice config for application name '%s'", application)
				continue
			}
			app, err := entities.CreateApplication(parameters.CreateApplicationParameters{
				Name:          application,
				PersianName:   appConfig.PersianName,
				NeedToApprove: appConfig.NeedToApprove,
				Branch:        appConfig.Branch,
			})
			if err != nil {
				log.Printf(err.Error())
			} else {
				applications = append(applications, app)
			}
		}
	}
	return applications
}

// create approvers and requester based on configuration and finally create job object with applications
func createJobWithCorrespondingApproversAndRequester(update *models.Update, applications []entities.Application) (entities.Job, error) {

	approvers, err := createApproversBasedOnConfiguration()
	if err != nil {
		return entities.Job{}, err
	}

	requester, err := createRequesterBasedOnUserMessage(update)
	if err != nil {
		return entities.Job{}, err
	}

	job, err := createJobBasedOnUserMessage(update)
	if err != nil {
		return entities.Job{}, err
	}

	if appErr := job.AddApplications(applications); appErr != nil {
		return entities.Job{}, appErr
	}

	if appErr := job.AddApprovers(approvers); appErr != nil {
		return entities.Job{}, appErr
	}

	if appErr := job.AddRequester(&requester); appErr != nil {
		return entities.Job{}, appErr
	}

	return job, nil
}

func createApproversBasedOnConfiguration() ([]entities.Approver, error) {
	config := configs.LoadConfig(configs.ConfigPath)
	var approvers []entities.Approver
	for _, approver := range config.Approvers {
		a, err := entities.CreateApprover(parameters.CreateApproverParameters{
			Username: approver.Username,
			FullName: approver.FullName,
		})
		if err == nil {
			approvers = append(approvers, a)
		}
	}
	if len(approvers) == 0 {
		return []entities.Approver{}, errors.New("there is no any corresponding approvers")
	}
	return approvers, nil
}

func createRequesterBasedOnUserMessage(update *models.Update) (entities.Requester, error) {
	requester, err := entities.CreateRequester(parameters.CreateRequesterParameters{
		Username: update.Message.From.Username,
		FullName: update.Message.From.FirstName + " " + update.Message.From.LastName,
	})
	if err != nil {
		return entities.Requester{}, errors.New("there is no any corresponding requester")
	}
	return requester, nil
}

func createJobBasedOnUserMessage(update *models.Update) (entities.Job, error) {
	job, err := entities.CreateJob(parameters.CreateJobParameters{
		ChatId: update.Message.Chat.ID,
	})
	if err != nil {
		return entities.Job{}, err
	}
	return job, nil
}
