# Multi-stage build for ultra-minimal RAM footprint and small image (<25MB)
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git and certs
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/mangahost ./cmd/server

# Final lightweight stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/mangahost /app/mangahost

EXPOSE 8080

CMD ["/app/mangahost"]
