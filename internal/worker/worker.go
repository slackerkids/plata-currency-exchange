package worker

import (
	"context"
	"errors"
	"log/slog"

	"github.com/slackerkids/plata-currency-exchange.git/internal/model"
	"github.com/slackerkids/plata-currency-exchange.git/internal/service"
)

var ErrJobQueueIsFull = errors.New("job queue is full")

type QuoteRepository interface {
	SetRate(ctx context.Context, result *model.QuoteResult) error
	SetStatusFailed(ctx context.Context, job *service.Job) error
}

type ExchangeRateClient interface {
	GetQuoteByCurrencyCode(ctx context.Context, base, quote string) (*model.QuoteResult, error)
}

type Pool struct {
	queue              chan service.Job
	quoteRepository    QuoteRepository
	exchangeRateClient ExchangeRateClient
}

func NewPool(
	ctx context.Context,
	numWorker int,
	quoteRepository QuoteRepository,
	exchangeRateClient ExchangeRateClient,
) *Pool {
	p := &Pool{
		queue:              make(chan service.Job, numWorker),
		quoteRepository:    quoteRepository,
		exchangeRateClient: exchangeRateClient,
	}

	for range numWorker {
		go p.worker(ctx)
	}

	return p
}

func (w *Pool) worker(ctx context.Context) {
	// TODO: Think about if database will fail and get quote by currency code will fail
	// 1. Backoff Retry
	// 2. If retries not helped then we should finish and return error up to main and close.
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-w.queue:
			if !ok {
				return
			}

			result, err := w.exchangeRateClient.GetQuoteByCurrencyCode(ctx, job.Base, job.Quote)
			if err != nil {
				slog.ErrorContext(ctx, "fetching external api", "error", err)

				if err := w.quoteRepository.SetStatusFailed(ctx, &job); err != nil {
					slog.ErrorContext(ctx, "set status to db", "error", err, "id", job.ID)
				}

				continue
			}

			if err := w.quoteRepository.SetRate(ctx, result); err != nil {
				slog.ErrorContext(ctx, "set rate to db", "error", err, "id", job.ID)
				continue
			}
		}

	}
}

func (w *Pool) AddJob(job *service.Job) error {
	var err error

	select {
	case w.queue <- *job:
	default:
		err = ErrJobQueueIsFull
	}

	return err
}
