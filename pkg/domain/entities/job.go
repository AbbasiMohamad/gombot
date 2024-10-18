package entities

import (
	"errors"
	"gombot/pkg/domain/dtos/parameters"
	"time"
)

type JobStatus string

const (
	Requested      JobStatus = "REQUESTED"
	NeedToApproved JobStatus = "NEED_TO_APPROVED"
	Confirmed      JobStatus = "CONFIRMED"
	InProgress     JobStatus = "IN_PROGRESS"
	Done           JobStatus = "DONE"
	Finished       JobStatus = "FINISHED"
	Canceled       JobStatus = "CANCELED"
	None           JobStatus = "NONE"
)

type Job struct {
	ID               uint
	ChatId           int64
	RequestMessageID int
	StatusMessageID  int
	Applications     []Application
	Approvers        []Approver
	Requester        *Requester
	Status           JobStatus
	CreatedAt        time.Time
}

// CreateJob creates new Job with status None or return error
func CreateJob(p parameters.CreateJobParameters) (Job, error) {
	if p.ChatId == 0 {
		return Job{}, errors.New("ChatId is not valid")
	}
	job := Job{
		ChatId:    p.ChatId,
		CreatedAt: time.Now(),
	}
	job.updateJobStatus()
	return job, nil
}

// updateJobStatus update status of Job based on job's circumstances.
// returns None if Applications, Approvers, or Requester be nil.
// returns Requested if Applications, Approvers, and Requester have valid data.
func (j *Job) updateJobStatus() {
	if j.Applications == nil || j.Approvers == nil || j.Requester == nil {
		j.Status = None
	}
	if j.Applications != nil || j.Approvers != nil || j.Requester != nil {
		j.Status = Requested
	}
}

func (j *Job) AddApplications(applications []Application) error {
	if applications == nil || len(applications) == 0 {
		return errors.New("applications is empty")
	}
	j.Applications = applications
	j.updateJobStatus()
	return nil
}

func (j *Job) AddApprovers(approvers []Approver) error {
	if approvers == nil || len(approvers) == 0 {
		return errors.New("approvers is empty")
	}
	j.Approvers = approvers
	j.updateJobStatus()
	return nil
}

func (j *Job) AddRequester(requester *Requester) error {
	if requester == nil {
		return errors.New("requester is empty")
	}
	j.Requester = requester
	j.updateJobStatus()
	return nil
}
