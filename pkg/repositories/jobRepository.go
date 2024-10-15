package repositories

import (
	"gombot/pkg/entities"
)

func InsertJob(job *entities.Job) uint {
	db := DbConnect()
	db.Create(&job)
	return job.ID
}

func GetJobByMessageId(messageId int) *entities.Job {
	db := DbConnect()
	var job *entities.Job
	db.Preload("Applications").Preload("Approvers").
		Preload("Requester").First(&job, "message_id = ?", messageId)
	return job
}

func GetRequestedJobs() []*entities.Job {
	db := DbConnect()
	var jobs []*entities.Job
	db.Preload("Applications").Preload("Approvers").
		Preload("Requester").Find(&jobs, "status = ?", entities.Requested)
	return jobs
}

func UpdateJob(job *entities.Job) {
	db := DbConnect()
	db.Model(&job).Updates(job)
}
