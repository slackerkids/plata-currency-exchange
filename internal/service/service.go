package service

import (
	"context"
	"fmt"
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

type QuoteRepository interface {
	GetLatestQuote(ctx context.Context, base string, quote string) (*Job, error)
	CreateQuote(ctx context.Context, job *Job) error
}

type Service struct {
	jobQueue        JobQueue
	quoteRepository QuoteRepository
}

func New(jobQueue JobQueue, quoteRepository QuoteRepository) *Service {
	return &Service{
		jobQueue:        jobQueue,
		quoteRepository: quoteRepository,
	}
}

func (s *Service) UpdateQuote(ctx context.Context, base, quote string) (*Job, error) {
	// 1. Try to create new job
	job := &Job{
		ID:    uuid.NewV7(),
		Base:  base,
		Quote: quote,
	}

	err := s.quoteRepository.CreateQuote(ctx, job)
	if err == nil {
		if err := s.jobQueue.AddJob(job); err != nil {
			return nil, fmt.Errorf("adding job: %w", err)
		}

		return job, nil
	}

	// 2. If job exists in db return existing job
	existingJob, err := s.quoteRepository.GetLatestQuote(ctx, base, quote)
	if err != nil {
		return nil, fmt.Errorf("get latest job: %w", err)
	}

	return existingJob, nil
}
