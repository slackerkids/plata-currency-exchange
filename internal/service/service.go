package service

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"uuid"

	"github.com/slackerkids/plata-currency-exchange.git/internal/model"
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
	GetOrCreateJob(ctx context.Context, job *Job) (*Job, error)
	GetQuoteByID(ctx context.Context, id uuid.UUID) (*model.QuoteResult, error)
	GetLatestQuote(ctx context.Context, base, quote string) (*model.QuoteResult, error)
	ListPendingJobs(ctx context.Context) (iter.Seq2[*Job, error], error)
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
	job := &Job{
		ID:    uuid.NewV7(),
		Base:  base,
		Quote: quote,
	}

	createdJob, err := s.quoteRepository.GetOrCreateJob(ctx, job)
	if err != nil {
		return nil, fmt.Errorf("create or get job: %w", err)
	}

	if job.ID != createdJob.ID {
		return createdJob, nil
	}

	if err := s.jobQueue.AddJob(job); err != nil {
		return nil, fmt.Errorf("adding job: %w", err)
	}

	return job, nil
}

func (s *Service) GetQuoteByID(ctx context.Context, id uuid.UUID) (*model.QuoteResult, error) {
	result, err := s.quoteRepository.GetQuoteByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get quote by id: %w", err)
	}

	return result, nil
}

func (s *Service) GetLatestQuote(ctx context.Context, base, quote string) (*model.QuoteResult, error) {
	result, err := s.quoteRepository.GetLatestQuote(ctx, base, quote)
	if err != nil {
		return nil, fmt.Errorf("get latest quote: %w", err)
	}

	return result, nil
}

func (s *Service) RecoverJobs(ctx context.Context) error {
	jobsList, err := s.quoteRepository.ListPendingJobs(ctx)
	if err != nil {
		return fmt.Errorf("listing pending jobs: %w", err)
	}

	for job, err := range jobsList {
		if err != nil {
			return err
		}

		s.jobQueue.AddJob(job)
	}

	slog.InfoContext(ctx, "jobs recovery started")
	return nil
}
