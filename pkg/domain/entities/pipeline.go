package entities

import "time"

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
