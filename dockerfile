# ========================
# Stage 1: Build Go binary
# ========================
FROM golang:1.25-alpine AS builder

# Install git for module downloads
RUN apk add --no-cache git

WORKDIR /app

# Copy go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go binary
RUN go build -o url-shortener ./cmd/server/main.go

# ========================
# Stage 2: Final lightweight image
# ========================
FROM alpine:latest

WORKDIR /app

# Copy Go binary from builder
COPY --from=builder /app/url-shortener . 

# Copy frontend assets
COPY --from=builder /app/frontend ./frontend

# Expose application port
EXPOSE 8080


# Run the Go binary
CMD ["./url-shortener"]
