# Стадия сборки
FROM golang:1.21-alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY app/go.mod app/go.sum ./
RUN go mod download

COPY app/*.go ./

RUN go build -o main .

# Стадия запуска
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8080

ENV PORT=8080

CMD ["./main"]
