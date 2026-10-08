FROM golang:1.27.1-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o /app/server ./cmd/main.go 

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

RUN adduser -D -u 10001 appuser
USER 10001:10001

WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

ENTRYPOINT ["./server"]