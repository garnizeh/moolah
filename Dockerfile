# Stage 1: Build
FROM golang:1.26-alpine AS builder

# Set the working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev curl

# Install Tailwind CLI (Required for Web build)
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v4.1.8/tailwindcss-linux-x64 && \
    chmod +x tailwindcss-linux-x64 && \
    mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss

# Install Templ CLI (Required for Web build)
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1001

# Copy go mod and sum files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Generate code (templ, etc)
RUN templ generate

# Build API binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o moolah-api ./cmd/api

# Build Web binary (Requires Tailwind to trigger CSS build if linked)
# Note: make web-build typically orchestrates this, but we run raw go build here
# to keep the Dockerfile self-contained or use the Makefile if preferred.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o moolah-web ./cmd/web

# Stage 2: API Final image
FROM alpine:3.20 AS api
RUN adduser -D moolah
USER moolah
WORKDIR /app
COPY --from=builder /app/moolah-api .
EXPOSE 8080
CMD ["./moolah-api"]

# Stage 3: Web Final image
FROM alpine:3.20 AS web
RUN adduser -D moolah
USER moolah
WORKDIR /app
COPY --from=builder /app/moolah-web .
EXPOSE 8081
CMD ["./moolah-web"]
