package entities

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

type JobStatus int

const (
	Requested      JobStatus = 1
	NeedToApproved JobStatus = 2
	Confirmed      JobStatus = 3 // TODO: study about isolation in go
	InProgress     JobStatus = 4 // TODO: study about isolation in go
	Done           JobStatus = 5
	Finished       JobStatus = 6
	Canceled       JobStatus = 7
	None           JobStatus = 8
)

// TODO: need to make isolate this module

var Queue []Job

type Job struct {
	ID           uint
	JobId        uuid.UUID
	ChatId       int64
	MessageId    int
	Applications []Application
	Approvers    []Approver
	Requester    Requester
	Status       JobStatus
	CreatedAt    time.Time
}

// TODO: fix approver & requester struct place in app
type Approver struct {
	ID         uint
	JobId      uint
	Username   string
	FullName   string //TODO: we dont have fullname of approvers
	IsApproved bool
}

type Requester struct {
	ID       uint
	JobId    uint
	Username string
	FullName string
}

func PushToQueue(q *Job) {
	if Queue == nil {
		Queue = make([]Job, 0)
	}
	Queue = append(Queue, *q)
}

func PopLastItemFromQueue() (*Job, error) {
	if Queue == nil || len(Queue) == 0 {
		return &Job{}, errors.New("There is no item in queue")
	}
	return &Queue[0], nil
}

func PopJobByChatIdFromQueue(chatId int64) (*Job, error) {
	if Queue == nil || len(Queue) == 0 {
		return &Job{}, errors.New("There is no item in queue")
	}
	for i, _ := range Queue {
		if Queue[i].ChatId == chatId {
			return &Queue[i], nil
		}
	}
	return &Job{}, errors.New("there is no item in queue")
}

func PopJobByMessageIdFromQueue(messageId int) (*Job, error) {
	if Queue == nil || len(Queue) == 0 {
		return &Job{}, errors.New("There is no item in queue")
	}
	for i, _ := range Queue {
		if Queue[i].MessageId == messageId {
			return &Queue[i], nil
		}
	}
	return &Job{}, errors.New("there is no item in queue")
}

func PopRequestedJobsFromQueue() ([]*Job, error) {
	if Queue == nil {
		return []*Job{}, errors.New("There is no item in queue")
	}
	var requestedJobs []*Job
	for i, _ := range Queue {
		if Queue[i].Status == Requested {
			requestedJobs = append(requestedJobs, &Queue[i])
		}
	}
	return requestedJobs, nil
}

func DequeueLastItemFromQueue() error {
	if Queue == nil {
		return errors.New("There is no item in queue")
	}
	Queue = Queue[1:]
	return nil
}
