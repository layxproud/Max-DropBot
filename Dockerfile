# ====== BUILD STAGE ======
FROM golang:1.26 AS builder

WORKDIR /app

# Кэшируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем код
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bot ./cmd/bot

# ====== RUNTIME STAGE ======
FROM alpine:3.20

WORKDIR /app

# Добавим CA сертификаты (важно для HTTPS!)
RUN apk add --no-cache ca-certificates

# Копируем бинарник
COPY --from=builder /app/bot .

# Запуск
CMD ["./bot"]