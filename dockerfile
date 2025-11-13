FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o reviewer-service ./cmd/reviewer-service

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.16.0

# Финальный образ
FROM alpine:latest
RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /root/

COPY --from=builder /app/reviewer-service .
COPY --from=builder /app/migrations ./migrations/
COPY --from=builder /go/bin/migrate /usr/local/bin/

EXPOSE 8080
CMD ["./reviewer-service"]