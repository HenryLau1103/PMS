# Simple build for Go backend
FROM golang:1.21-alpine

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy all files first
COPY . .

# Generate go.sum and download dependencies
RUN go mod tidy && go mod download

# Build the application
RUN go build -o psm-backend ./cmd/api

# Expose port
EXPOSE 8080

# Start the application
CMD ["./psm-backend"]
