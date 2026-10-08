package handler

import (
	"context"
	"encoding/json/v2"
	"log/slog"
	"net/http"
	"uuid"

	"github.com/slackerkids/plata-currency-exchange.git/internal/model"
	"github.com/slackerkids/plata-currency-exchange.git/internal/service"
)

type Service interface {
	UpdateQuote(ctx context.Context, base, quote string) (*service.Job, error)
	GetQuoteByID(ctx context.Context, id uuid.UUID) (*model.QuoteResult, error)
	GetLatestQuote(ctx context.Context, base, quote string) (*model.QuoteResult, error)
}

type Handler struct {
	service Service
}

type UpdateQuoteRequest struct {
	BaseCurrency  string `json:"base_currency"`
	QuoteCurrency string `json:"quote_currency"`
}

func New(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) UpdateQuote(w http.ResponseWriter, r *http.Request) {
	var req UpdateQuoteRequest
	if err := json.UnmarshalRead(r.Body, &req); err != nil {
		http.Error(w, "request body incorrect", http.StatusBadRequest)
		return
	}

	job, err := h.service.UpdateQuote(r.Context(), req.BaseCurrency, req.QuoteCurrency)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.MarshalWrite(w, job); err != nil {
		slog.ErrorContext(r.Context(), "marshal write", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetQuoteByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.service.GetQuoteByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.MarshalWrite(w, result); err != nil {
		slog.ErrorContext(r.Context(), "marshal write", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetLatestValueByCurrencyCode(w http.ResponseWriter, r *http.Request) {
	base := r.URL.Query().Get("base")
	quote := r.URL.Query().Get("quote")

	if base == "" && quote == "" {
		http.Error(w, "base and quote required", http.StatusBadRequest)
		return
	}

	result, err := h.service.GetLatestQuote(r.Context(), base, quote)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.MarshalWrite(w, result); err != nil {
		slog.ErrorContext(r.Context(), "marshal write", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
