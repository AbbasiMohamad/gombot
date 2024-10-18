package entities

import (
	"errors"
	"gombot/pkg/domain/dtos/parameters"
	"log"
	"time"
)

type Pipeline struct {
	ID            uint
	ApplicationID uint
	PipelineID    int
	Status        string
	Ref           string
	WebURL        string
	CreatedAt     time.Time
	FinishedAt    time.Time
}

func CreatePipeline(p parameters.CreatePipelineParameters) (Pipeline, error) {
	if p.Status == "" || p.Ref == "" || p.PipelineID == 0 {
		log.Printf("can not create pipeline with empty parameters.")
		return Pipeline{}, errors.New("can not create pipeline with with empty parameters")
	}
	pipeline := Pipeline{
		Status:     p.Status,
		Ref:        p.Ref,
		CreatedAt:  time.Now(),
		PipelineID: p.PipelineID,
		WebURL:     p.WebURL,
	}
	log.Printf("created pipeline with id '%d' and make its status '%s'.", pipeline.ID, pipeline.Status)
	return pipeline, nil
}
