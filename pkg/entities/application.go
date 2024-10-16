package entities

import "time"

type Application struct {
	ID              uint
	JobID           uint
	Name            string
	PersianName     string
	Status          ApplicationStatus
	GitlabUrl       string
	Branch          string
	NeedToApprove   bool
	Pipeline        Pipeline
	GitlabProjectID int
}

type Pipeline struct {
	ID            uint
	ApplicationID uint
	PipelineID    int
	Status        string
	Ref           string
	WebURL        string
	CreatedAt     time.Time
}

type ApplicationStatus string // TODO: GPT this statement

const (
	Declared   ApplicationStatus = "declared"
	Pending    ApplicationStatus = "pending"
	Processing ApplicationStatus = "processing"
	Failed     ApplicationStatus = "deployfailed"
	Deployed   ApplicationStatus = "deployed"
)
