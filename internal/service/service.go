package service

import (
	"context"

	"github.com/slackerkids/plata-currency-exchange.git/internal/model"
)

type ExchangeRateClient interface {
	GetQuoteByCurrencyCode(ctx context.Context, base, quote string) (*model.QuoteResult, error)
}
