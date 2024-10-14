package models

import (
	"errors"
	"github.com/google/uuid"
)

// TODO: need to make isolate this module
var Queue []Job

type Job struct {
	JobId        uuid.UUID
	ChatId       int64
	MessageId    int
	Applications []Application
	Approvers    []Approver
	Requester    Requester
	Status       JobStatus
}
type JobStatus string

const (
	Requested      JobStatus = "requested"
	NeedToApproved JobStatus = "needtoapproved"
	Confirmed      JobStatus = "confirmed" // TODO: study about isolation in go
	Done           JobStatus = "approved"
)

// TODO: fix approver & requester struct place in app
type Approver struct {
	Username string
}

type Requester struct {
	Username string
}

func PushToQueue(q *Job) {
	if Queue == nil {
		Queue = make([]Job, 0)
	}
	Queue = append(Queue, *q)
}

func PopFromQueue() (*Job, error) {
	if Queue == nil {
		return &Job{}, errors.New("There is no item in queue")
	}
	return &Queue[0], nil
}
