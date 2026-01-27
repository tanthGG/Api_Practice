# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=go1.24.4 

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o marketplace-merchant-api ./cmd/main.go

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/marketplace-merchant-api .
COPY --from=builder /app/config ./config

# Expose port
EXPOSE 80

# Run the application
CMD ["./marketplace-merchant-api"]

