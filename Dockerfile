# Stage 1: Build
FROM golang:1.26-alpine AS builder

# Set the working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# Using -ldflags="-s -w" to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o moolah-api ./cmd/api

# Stage 2: Final image
FROM alpine:3.20

# Add non-root user for security
RUN adduser -D moolah
USER moolah

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/moolah-api .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./moolah-api"]
