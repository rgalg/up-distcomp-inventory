FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install ca-certificates to handle SSL issues
RUN apk --no-cache add ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates wget
WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8003

CMD ["./main"]