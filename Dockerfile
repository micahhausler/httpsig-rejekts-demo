# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

ENV GOPROXY=direct

# Download dependencies
RUN go mod download

# Copy source code
COPY pkg pkg
COPY cmd cmd

# Build server binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server cmd/server/main.go

# Final stage for server
# FROM gcr.io/distroless/static-debian12
FROM public.ecr.aws/eks-distro-build-tooling/eks-distro-minimal-base:latest-al23

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/bin/server /app/server

# Expose the port the server runs on
EXPOSE 8080

# Run the server
ENTRYPOINT ["/app/server"] 