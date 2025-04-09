FROM golang:1.23.4-alpine AS builder

WORKDIR /app

RUN go mod download

COPY cmd/app .

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN apk add --no-cache git make && \
    go install github.com/swaggo/swag/cmd/swag@v1.16.3

COPY go.mod ./
COPY go.sum ./

RUN mkdir -p docs/swagger

RUN swag init \
    --parseDependency \
    --parseInternal \
    --parseDepth 5 \
    -g internal/app/start/start.go \
    --output docs/swagger

# Сборка приложения
RUN go build -mod=vendor -ldflags="-s -w" -o ./bin/app ./cmd/app/*

FROM alpine:latest
WORKDIR cmd/app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /app/bin/app /app/
COPY --from=builder /app/docs/swagger /app/docs/swagger

WORKDIR /app

# Переключение на непривилегированного пользователя
USER nobody

# Открытие порта
EXPOSE 8080

# Команда запуска приложения
CMD ["/app/app"]

#diagnose id
#5160103D-098E-43EE-A379-37F8D15BB952/20250409143423
