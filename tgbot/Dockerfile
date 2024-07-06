FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN apk add build-base && apk cache clean
ENV CGO_ENABLED=1
RUN go build -o ./tgbot ./cmd/tgbot/tgbot.go
FROM alpine
WORKDIR /app
COPY --from=builder /app/tgbot ./tgbot
EXPOSE 8080
CMD ["./tgbot"]