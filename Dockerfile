# Use Go for building
FROM golang:1.20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy files
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
RUN go build -o app .

# Use a lightweight image for running
FROM alpine:latest

# Set working directory
WORKDIR /

# Copy compiled binary
COPY --from=builder /app/app /app

# Set execution permissions
RUN chmod +x /app

# Run the application
CMD ["/app", "s", "ws://app:3000"]
