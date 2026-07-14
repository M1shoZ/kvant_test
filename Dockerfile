FROM golang:1.26-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей и скачиваем их отдельно
# Это ускорит последующие сборки, если зависимости не менялись
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/app/main.go

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]