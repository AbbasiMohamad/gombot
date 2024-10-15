package repositories

import (
	"github.com/google/uuid"
	"gombot/pkg/entities"
)

func InsertJob(job *entities.Job) uint {
	db := DbConnect()
	db.Create(&job)
	return job.ID
}

func GetJobById(jobId uuid.UUID) *entities.Job {
	db := DbConnect()
	var job *entities.Job
	db.First(&job, "job_id = ?", jobId)
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
