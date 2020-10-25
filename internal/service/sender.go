package service

import "projects/email-sending-service/internal/repository"

type Sender interface {
	Send() error
}

type sender struct {
	r repository.EmailRepository
}

