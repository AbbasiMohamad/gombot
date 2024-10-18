package entities

import (
	"errors"
	"gombot/pkg/domain/dtos/parameters"
	"log"
)

type Approver struct {
	ID         uint
	JobId      uint
	Username   string
	FullName   string
	IsApproved bool
}

// CreateApprover creates valid Approver or return error
func CreateApprover(p parameters.CreateApproverParameters) (Approver, error) {
	if p.Username == "" || p.FullName == "" {
		log.Printf("can not create approver with empty username or full name")
		return Approver{}, errors.New("can not create approver with empty username or full name")
	}
	approver := Approver{
		Username: p.Username,
		FullName: p.FullName,
	}
	log.Printf("created approver with fullname of '%s'", approver.FullName)
	return approver, nil
}
