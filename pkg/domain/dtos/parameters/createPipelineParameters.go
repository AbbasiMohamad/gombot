package parameters

import "time"

type CreatePipelineParameters struct {
	PipelineID int
	Status     string
	Ref        string
	WebURL     string
	CreatedAt  time.Time
}
