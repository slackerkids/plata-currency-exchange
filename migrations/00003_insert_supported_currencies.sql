-- +goose Up
INSERT INTO supported_currencies (name)
VALUES ('EUR'), ('USD'), ('MXN'), ('AUD'), ('KZT')
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM supported_currencies;  