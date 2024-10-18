package entities

import (
	"errors"
	"gombot/pkg/domain/dtos/parameters"
	"log"
)

type ApplicationStatus string // TODO: GPT this statement

const (
	Declared   ApplicationStatus = "DECLARED"
	Pending    ApplicationStatus = "PENDING"
	Processing ApplicationStatus = "PROCESSING"
	Failed     ApplicationStatus = "FAILED"
	Deployed   ApplicationStatus = "DEPLOYED"
)

type Application struct {
	ID              uint
	JobID           uint
	Name            string
	PersianName     string
	GitlabUrl       string
	GitlabProjectID int
	Branch          string
	NeedToApprove   bool
	Status          ApplicationStatus
	Pipeline        Pipeline
}

// CreateApplication creates new Application with status Declared or return error
func CreateApplication(p parameters.CreateApplicationParameters) (Application, error) {
	if p.Name == "" || p.PersianName == "" || p.Branch == "" {
		log.Printf("can not create application with empty parameters.")
		return Application{}, errors.New("can not create application with with empty parameters")
	}
	app := Application{
		Name:          p.Name,
		PersianName:   p.PersianName,
		Branch:        p.Branch,
		NeedToApprove: p.NeedToApprove,
		Status:        Declared,
	}
	log.Printf("created application named '%s' and make its status '%s' ", app.Name, app.Status)
	return app, nil
}

func (a *Application) SetStatusToPending() {
	a.Status = Pending
}
