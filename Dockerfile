FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o muzz-explore-service ./cmd/server

# Install migrate
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/muzz-explore-service .
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY internal/db/migrations ./internal/db/migrations

EXPOSE 8080
CMD ["./muzz-explore-service"]