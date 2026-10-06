-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX currencies_pending_unique_idx
ON quote (base_currency, quote_currency)
WHERE status = 'PENDING';
-- +goose StatementEnd

-- +goose Down