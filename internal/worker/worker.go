package worker

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/slackerkids/plata-currency-exchange.git/internal/model"
	"github.com/slackerkids/plata-currency-exchange.git/internal/service"
)

var ErrJobQueueIsFull = errors.New("job queue is full")

type QuoteRepository interface {
	SetRate(ctx context.Context, result *model.QuoteResult) error
	SetStatusFailed(ctx context.Context, job *service.Job) error
}

type ExchangeRateClient interface {
	GetQuoteByCurrencyCode(ctx context.Context, job *service.Job) (*model.QuoteResult, error)
}

type Pool struct {
	queue              chan service.Job
	quoteRepository    QuoteRepository
	exchangeRateClient ExchangeRateClient
	wg                 sync.WaitGroup
}

func NewPool(
	ctx context.Context,
	numWorker int,
	queueSize int,
	quoteRepository QuoteRepository,
	exchangeRateClient ExchangeRateClient,
) *Pool {
	p := &Pool{
		queue:              make(chan service.Job, queueSize),
		quoteRepository:    quoteRepository,
		exchangeRateClient: exchangeRateClient,
		wg:                 sync.WaitGroup{},
	}

	for range numWorker {
		p.wg.Go(func() {
			p.worker(ctx)
		})
	}
	return p
}

func (w *Pool) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-w.queue:
			if !ok {
				return
			}

			result, err := w.exchangeRateClient.GetQuoteByCurrencyCode(ctx, &job)
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

func (w *Pool) Close(ctx context.Context) error {
	jobsDone := make(chan struct{})

	go func() {
		close(w.queue)
		w.wg.Wait()
		close(jobsDone)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-jobsDone:
		return nil
	}
}
