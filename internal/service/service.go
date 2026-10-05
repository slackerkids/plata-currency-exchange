package service

import (
	"uuid"
)

type Job struct {
	ID    uuid.UUID
	Base  string
	Quote string
}

type JobQueue interface {
	AddJob(job Job) bool
}

type Service struct {
}
