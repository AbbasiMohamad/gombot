package models

type Application struct {
	Name   string
	Status ApplicationStatus
}

type ApplicationStatus string

const (
	Declared ApplicationStatus = "declared"
)
