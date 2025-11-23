FROM golang:1.25.3-alpine AS builder

WORKDIR /app

# Install certificates for Go module downloads
RUN apk --no-cache add ca-certificates

# Use direct proxy to avoid certificate issues
ENV GOPROXY=direct

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN apk add --no-cache git
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd

FROM alpine:latest
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

EXPOSE 8001 9001

CMD ["./main"]