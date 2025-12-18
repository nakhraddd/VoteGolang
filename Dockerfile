FROM golang:1.24.0-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/vote-api ./cmd/app/main.go

FROM alpine:3.21

WORKDIR /app
COPY --from=builder /app/vote-api /app/vote-api

CMD ["/app/vote-api"]