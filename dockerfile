FROM golang:1.24.2 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -a -ldflags "-s -w" -o main ./cmd/bot/main.go

FROM debian:stable-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/main .
WORKDIR /app/data
ENV PORT=8080
EXPOSE 8080
CMD ["/app/main"]
