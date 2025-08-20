FROM golang:1.23.8-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o app .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app .
COPY .env .

RUN mkdir -p /app/logs

EXPOSE 5555
CMD ["./app"]
