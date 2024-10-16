package entities

import (
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
	ID               uint
	JobId            uuid.UUID
	ChatId           int64
	RequestMessageID int
	StatusMessageID  int
	Applications     []Application
	Approvers        []Approver
	Requester        Requester
	Status           JobStatus
	CreatedAt        time.Time
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
