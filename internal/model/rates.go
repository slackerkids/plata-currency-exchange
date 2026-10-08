package model

import (
	"time"
	"uuid"
)

type QuoteResult struct {
	ID         uuid.UUID
	Base       string
	Quote      string
	Rate       *float64
	Status     string
	FinishedAt *time.Time
	CreatedAt  *time.Time
}
