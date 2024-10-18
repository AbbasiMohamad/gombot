package repositories

import (
	"errors"
	"gombot/pkg/domain/entities"
	"gorm.io/gorm"
)

func InsertJob(job *entities.Job) uint {
	db = DbConnect()
	db.Create(&job)
	return job.ID
}

func GetJobByMessageId(messageId int) *entities.Job {
	db = DbConnect()
	var job *entities.Job
	db.Preload("Applications").Preload("Approvers").
		Preload("Requester").First(&job, "request_message_id = ?", messageId)
	return job
}

func GetRequestedJobs() []*entities.Job {
	db = DbConnect()
	var jobs []*entities.Job
	db.Preload("Applications").Preload("Approvers").
		Preload("Requester").Find(&jobs, "status = ?", entities.Requested)
	return jobs
}

func UpdateJobOld(job *entities.Job) {
	db = DbConnect()
	db.Transaction(func(tx *gorm.DB) error {
		// First update the Job
		if err := tx.Save(&job).Error; err != nil {
			return err
		}

		// Now update the Approvers
		for _, approver := range job.Approvers {
			if err := tx.Model(&entities.Approver{}).Where("id = ?", approver.ID).Updates(approver).Error; err != nil { //TODO: investigate this syntax
				return err
			}
		}

		// Now update the Application
		for _, app := range job.Applications {
			if err := tx.Model(&entities.Application{}).Where("id = ?", app.ID).Updates(app).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func UpdateJob(job *entities.Job) {
	db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&job)
}

func GetFirstConfirmedJob() (*entities.Job, error) {
	db = DbConnect()
	var job *entities.Job
	result := db.Preload("Applications").
		Preload("Approvers").
		Preload("Requester").
		Where("status = ?", entities.Confirmed).
		Order("created_at").First(&job)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("there is no approved job in the database")
		}
		return nil, result.Error
	}
	return job, nil
}

func GetFirstInProgressJob() (*entities.Job, error) {
	db = DbConnect()
	var job *entities.Job
	result := db.Preload("Applications").Preload("Applications.Pipeline").
		Preload("Approvers").
		Preload("Requester").
		Where("status = ?", entities.InProgress).
		Order("created_at").First(&job)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("there is no in progress job in the database")
		}
		return nil, result.Error
	}

	return job, nil
}

func GetApplicationById(id uint) (*entities.Application, error) {
	db = DbConnect()
	var app *entities.Application
	result := db.Preload("Pipeline").
		Where("id = ?", id).
		First(&app)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("there is no in progress job in the database")
		}
		return nil, result.Error
	}

	return app, nil
}

func GetFirstDoneJob() (*entities.Job, error) {
	db = DbConnect()
	var job *entities.Job
	// Query the first approved job
	result := db.Preload("Applications").
		Preload("Approvers").
		Preload("Requester").
		Where("status = ?", entities.Done).
		Order("created_at").First(&job)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("there is no approved job in the database")
		}
		return nil, result.Error // Return any other database error
	}

	return job, nil
}

func IsInProgressJobExists() bool {
	db = DbConnect()
	var job *entities.Job
	db.Find(&job, "status = ?", entities.InProgress) // TODO : make query with count
	if job.ID == 0 {                                 // TODO: improve error checking
		return false
	}
	return true
}
