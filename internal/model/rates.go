package model

import "uuid"

var (
	StatusPending = StatusEnum{name: "PENDING"}
	StatusDone    = StatusEnum{name: "DONE"}
	StatusFail    = StatusEnum{name: "FAILED"}
)

type StatusEnum struct {
	name string
}

func (s StatusEnum) String() string {
	if s.name == "" {
		// Default is pending
		return "PENDING"
	}

	return s.name
}

type QuoteResult struct {
	ID        uuid.UUID
	Base      string
	Quote     string
	Rate      *float64
	Status    StatusEnum
	UpdatedAt *string
	CreatedAt *string
}
