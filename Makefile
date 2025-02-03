generate:
	buf generate proto/explore-service.proto
	sqlc generate

test:
	go test -v -race ./...

build:
	docker-compose build

run:
	docker-compose up

# Install development dependencies
deps:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Stop all containers
stop:
	docker-compose down
