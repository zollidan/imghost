FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy HTML templates and static files
COPY --from=builder /app/index.html .
COPY --from=builder /app/static ./static

# Create directories for uploads and encrypted files
RUN mkdir -p /app/uploads /app/encrypted

# Expose port
EXPOSE 8000

# Run the application
CMD ["./main"]
