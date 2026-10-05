package worker

import (
	"context"
	"log/slog"

	"github.com/slackerkids/plata-currency-exchange.git/internal/model"
	"github.com/slackerkids/plata-currency-exchange.git/internal/service"
)

type QuoteRepository interface {
	SetRate(result *model.QuoteResult)
	SetStatusFailed() // TODO: add params and return value
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
				w.quoteRepository.SetStatusFailed()
				continue
			}

			w.quoteRepository.SetRate(result)
		}

	}
}

func (w *Pool) AddJob(job service.Job) (queueFull bool) {
	select {
	case w.queue <- job:
	default:
		queueFull = true
	}

	return
}
