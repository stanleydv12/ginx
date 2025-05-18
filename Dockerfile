# Build stage
FROM golang:1.23.2-alpine AS builder

# Set working directory
WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Development stage
FROM builder AS development

# Install Air for hot-reloading
RUN go install github.com/air-verse/air@latest

# Copy only necessary files for development
COPY --from=builder /go/pkg /go/pkg
COPY go.mod go.sum ./

# Set the entrypoint to use Air
ENTRYPOINT ["air"]

# Production build stage
FROM builder AS production

# Build the application
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/bin/ginx ./cmd/...

# Final stage
FROM alpine:3.18

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=production /app/bin/ginx /usr/local/bin/ginx

# Copy configuration files
COPY --from=production /app/config ./config

# Create a non-root user
RUN adduser -D -g '' appuser
USER appuser

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["ginx"]
