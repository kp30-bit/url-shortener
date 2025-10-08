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

# Default environment variables (can be overridden at runtime)
ENV CONFIG_ENV=prod
ENV PORT=8080
ENV MONGO_URI="mongodb+srv://kamalpratik:youwillneverwalkalone@cluster0.lu5o0r2.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"
ENV MONGO_DB="Cluster0"
ENV HOST="kliplink"

# Run the Go binary
CMD ["./url-shortener"]
