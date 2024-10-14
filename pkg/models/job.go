package models

import "github.com/google/uuid"

var Queue []Job

type Job struct {
	JobId        uuid.UUID
	ChatId       int64
	MessageId    int
	Applications []Application
}

func PushToQueue(q *Job) {
	if Queue == nil {
		Queue = make([]Job, 0)
	}
	Queue = append(Queue, *q)
}
