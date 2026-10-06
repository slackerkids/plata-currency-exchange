all: clean postgres

run:
	go run cmd/main.go

build:
	go build -o bin/exchange-rate cmd/main.go

postgres:
	docker run -d \
		--name postgres \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=plata-currency-exchange \
		-p 5432:5432 \
		postgres

clean:
	docker rm -f postgres

db-connect:
	PGPASSWORD=postgres psql -h localhost -U postgres -p 5432 plata-currency-exchange