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

// Job is representation of a request of update. Job must execute serially but Application of it must run parallel.
type Job struct {
	ID               uint
	ChatId           int64
	RequestMessageID int
	StatusMessageID  int
	Applications     []Application
	Approvers        []Approver
	Requester        *Requester
	Status           JobStatus // TODO: make this isolate
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
// returns NeedToApproved if Status is Requested and RequestMessageID has a content.
// returns Confirmed if Status is NeedToApproved and all approvers approved.
// returns InProgress if Status is Confirmed and StatusMessageID has a content.
func (j *Job) updateJobStatus() {
	if j.Applications == nil || j.Approvers == nil || j.Requester == nil {
		j.Status = None
	}
	if j.Applications != nil || j.Approvers != nil || j.Requester != nil {
		j.Status = Requested
	}
	if j.RequestMessageID > 0 && j.Status == Requested {
		j.Status = NeedToApproved
	}
	if j.checkAllApproversApproved() && j.Status == NeedToApproved {
		j.Status = Confirmed
	}
	if j.StatusMessageID > 0 && j.Status == Confirmed {
		j.Status = InProgress
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

// SetRequestMessageID make job status to NeedToApproved.
func (j *Job) SetRequestMessageID(messageId int) error {
	if j.Status != Requested {
		return errors.New("job status is invalid for operation named 'SetRequestMessageID'")
	}
	j.RequestMessageID = messageId
	j.updateJobStatus()
	for i := range j.Applications {
		j.Applications[i].SetStatusToPending()
	}
	return nil
}

// SetApproverResponse can make job status to Confirmed.
func (j *Job) SetApproverResponse(username string) error {
	if j.Status != NeedToApproved {
		return errors.New("job status is invalid for operation named 'SetApproverResponse'")
	}
	for i := range j.Approvers {
		if j.Approvers[i].Username == username {
			j.Approvers[i].IsApproved = true
			return nil
		}
	}
	j.updateJobStatus()
	return errors.New("approver is not found")
}

func (j *Job) checkAllApproversApproved() bool {
	allApproved := true
	for _, approver := range j.Approvers {
		if !approver.IsApproved {
			allApproved = false
			break
		}
	}
	return allApproved
}

// SetStatusMessageID make job status to InProgress.
func (j *Job) SetStatusMessageID(messageId int) error {
	if j.Status != Confirmed {
		return errors.New("job status is invalid for operation named 'SetStatusMessageID'")
	}
	j.StatusMessageID = messageId
	j.updateJobStatus()
	for i := range j.Applications {
		j.Applications[i].SetStatusToProcessing()
	}
	return nil
}
