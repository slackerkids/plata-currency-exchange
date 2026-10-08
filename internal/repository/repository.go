package repository

import (
	"context"
	"fmt"
	"iter"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/slackerkids/plata-currency-exchange.git/internal/model"
	"github.com/slackerkids/plata-currency-exchange.git/internal/service"
)

const (
	setRateQuery = `
	UPDATE quote 
	SET rate = $1, status = $2, finished_at = $3 
	WHERE id = $4
	`

	setStatusFailedQuery = `
	UPDATE quote 
	SET status = $1, finished_at = $2 
	WHERE id = $3"
	`

	getLatestQuoteQuery = `
	SELECT id, rate, status, finished_at, created_at
	FROM quote 
	WHERE base_currency = $1 AND quote_currency = $2 AND status = 'DONE'
	ORDER BY finished_at DESC LIMIT 1
	`

	getOrCreateJobQuery = `
	INSERT INTO quote (id,base_currency, quote_currency)
	VALUES ($1, $2, $3)
	ON CONFLICT (base_currency, quote_currency) WHERE status = 'PENDING'
	DO UPDATE SET base_currency = EXCLUDED.base_currency
	RETURNING id;
	`

	getQuoteByIDQuery = `
	SELECT base_currency, quote_currency, rate, status, finished_at, created_at 
	FROM quote 
	WHERE id = $1
	`

	listPendingJobsQuery = `
	SELECT id, base_currency, quote_currency
	FROM quote
	WHERE status = 'PENDING'
	ORDER BY created_at ASC
	`
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) SetRate(ctx context.Context, result *model.QuoteResult) error {
	_, err := r.db.Exec(
		ctx,
		setRateQuery,
		result.Rate,
		result.Status,
		time.Now(),
		result.ID,
	)

	if err != nil {
		return fmt.Errorf("setting rate: %w", err)
	}

	return nil
}

func (r *Repository) SetStatusFailed(ctx context.Context, job *service.Job) error {
	_, err := r.db.Exec(
		ctx,
		setStatusFailedQuery,
		"FAILED",
		time.Now(),
		job.ID,
	)

	if err != nil {
		return fmt.Errorf("setting status fail: %w", err)
	}

	return nil
}

func (r *Repository) GetLatestQuote(ctx context.Context, base string, quote string) (*model.QuoteResult, error) {
	// Getting lasest quote by currency code
	quoteRes := &model.QuoteResult{
		Base:  base,
		Quote: quote,
	}

	row := r.db.QueryRow(
		ctx,
		getLatestQuoteQuery,
		base,
		quote,
	)

	if err := row.Scan(
		&quoteRes.ID,
		&quoteRes.Rate,
		&quoteRes.Status,
		&quoteRes.FinishedAt,
		&quoteRes.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("getting latest quote: %w", err)
	}

	return quoteRes, nil
}

func (r *Repository) GetOrCreateJob(ctx context.Context, job *service.Job) (*service.Job, error) {
	createdJob := &service.Job{
		Base:  job.Base,
		Quote: job.Quote,
	}

	row := r.db.QueryRow(
		ctx,
		getOrCreateJobQuery,
		job.ID,
		job.Base,
		job.Quote,
	)

	if err := row.Scan(&createdJob.ID); err != nil {
		return nil, fmt.Errorf("scanning created job: %w", err)
	}

	return createdJob, nil
}

func (r *Repository) GetQuoteByID(ctx context.Context, id uuid.UUID) (*model.QuoteResult, error) {
	quoteRes := &model.QuoteResult{
		ID: id,
	}

	row := r.db.QueryRow(
		ctx,
		getQuoteByIDQuery,
		id,
	)

	// base_currency, quote_currency, rate, status, finished_at, created_at
	if err := row.Scan(
		&quoteRes.Base,
		&quoteRes.Quote,
		&quoteRes.Rate,
		&quoteRes.Status,
		&quoteRes.FinishedAt,
		&quoteRes.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("getting latest quote: %w", err)
	}

	return quoteRes, nil
}

func (r *Repository) ListPendingJobs(ctx context.Context) (iter.Seq2[*service.Job, error], error) {
	rows, err := r.db.Query(
		ctx,
		listPendingJobsQuery,
	)
	if err != nil {
		return nil, err
	}

	return func(yield func(*service.Job, error) bool) {
		defer rows.Close()

		for rows.Next() {
			var job service.Job
			err := rows.Scan(&job.ID, &job.Base, &job.Quote)

			if !yield(&job, err) {
				break
			}
		}

	}, nil
}
