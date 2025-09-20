FROM golang:1.24 AS builder
WORKDIR /app
COPY . .
RUN go build -o server main.go

# bullseye → bookworm に変更
FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/server .
CMD ["./server"]
