# Build Stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy workspace files
COPY go.work .
COPY go.work.sum .
COPY backend/ ./backend/

# Build specific service (passed as build arg)
ARG SERVICE_PATH
RUN go build -o /app/service ./backend/services/${SERVICE_PATH}

# Run Stage
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/service .
COPY --from=builder /app/backend/pkg/common/config.go . 

# Expose port (can be overridden)
EXPOSE 8080

CMD ["./service"]
