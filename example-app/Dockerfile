# Simple Go HTTP server example application
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy all files
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/app .

# Expose port
EXPOSE 8080


# Start the application
CMD ["/app/app"]
