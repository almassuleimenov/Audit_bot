# Шаг 1: Сборка
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем файлы модулей и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем бинарник (отключаем CGO для минимального веса и кроссплатформенности)
RUN CGO_ENABLED=0 GOOS=linux go build -o audit_bot main.go

# Шаг 2: Финальный образ
FROM alpine:latest

WORKDIR /root/

# Добавляем таймзоны (важно для логов РК)
RUN apk --no-cache add tzdata
ENV TZ=Asia/Almaty

# Копируем собранный файл из первого шага
COPY --from=builder /app/audit_bot .
COPY --from=builder /app/.env . 

# Запускаем бота
CMD ["./audit_bot"]