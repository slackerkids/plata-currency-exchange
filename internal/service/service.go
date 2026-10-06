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
	AddJob(job *Job) error
}

type Service struct {
	jobQueue JobQueue
}
