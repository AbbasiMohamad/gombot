package entities

import (
	"errors"
	"gombot/pkg/domain/dtos/parameters"
	"log"
)

type Requester struct {
	ID       uint
	JobId    uint
	Username string
	FullName string
}

// CreateRequester creates valid Requester or return error
func CreateRequester(p parameters.CreateRequesterParameters) (Requester, error) {
	if p.Username == "" || p.FullName == "" {
		log.Printf("can not create approver with empty username or full name.")
		return Requester{}, errors.New("can not create requester with empty username or full name")
	}
	requester := Requester{
		Username: p.Username,
		FullName: p.FullName,
	}
	log.Printf("created requester with fullname of '%s'", requester.FullName)
	return requester, nil
}
