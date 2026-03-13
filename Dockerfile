# Step 1: Build the binary
FROM golang:1.25-alpine AS builder

# Set the working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# We use CGO_ENABLED=0 to create a statically linked binary (better for portability)
# GOOS=linux ensures it compiles for the target OS of the container
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go

# Step 2: Create the final lightweight image
FROM alpine:latest

# Install ca-certificates (needed for sending emails via SMTP/TLS)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Expose the gRPC port
EXPOSE 50051

# Command to run
CMD ["./main"]
