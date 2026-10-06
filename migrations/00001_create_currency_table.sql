-- +goose Up
-- +goose StatementBegin
CREATE TYPE status_enum AS ENUM ('PENDING', 'DONE', 'FAILED');

CREATE TABLE IF NOT EXISTS supported_currencies (
        name CHAR(3) PRIMARY KEY NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS quote (
        id UUID PRIMARY KEY NOT NULL,
        base_currency CHAR(3) REFERENCES supported_currencies(name) NOT NULL,
        quote_currency CHAR(3) REFERENCES supported_currencies(name) NOT NULL,
        rate DECIMAL,
        status status_enum NOT NULL DEFAULT 'PENDING',
        finished_at TIMESTAMP WITH TIME ZONE,
        created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX quote_base_quote_idx ON quote (base_currency, quote_currency);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS quote;
DROP TABLE IF EXISTS supported_currencies;
DROP TYPE IF EXISTS status_enum;
-- +goose StatementEnd
