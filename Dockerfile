FROM golang:1.23.8-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o app .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app .

RUN mkdir -p /app/logs

EXPOSE 4444
CMD ["./app"]
