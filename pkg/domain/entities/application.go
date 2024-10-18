package entities

import (
	"gombot/pkg/domain/dtos"
	"time"
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
	status          ApplicationStatus
	Pipeline        Pipeline
}

type Pipeline struct {
	ID            uint
	ApplicationID uint
	PipelineID    int
	Status        string
	Ref           string
	WebURL        string
	CreatedAt     time.Time
	FinishedAt    time.Time
}

func CreateApplication(appDto dtos.CreateApplicationDto) Application {
	return Application{}
}

type ApplicationStatus string // TODO: GPT this statement

const (
	Declared   ApplicationStatus = "declared"
	Pending    ApplicationStatus = "pending"
	Processing ApplicationStatus = "processing"
	Failed     ApplicationStatus = "failed"
	Deployed   ApplicationStatus = "deployed"
)
