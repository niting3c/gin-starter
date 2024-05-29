# Use a specific Go version tag for consistency and reproducibility
FROM ubuntu:latest AS builder
ENV TZ=UTC

# Install Golang and CA certificates
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates tzdata golang --no-install-recommends && \
    update-ca-certificates

# Set the working directory
WORKDIR /app

# Copy only go.mod and go.sum first for efficient caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the remaining source code
COPY . .

# Build the Go application
RUN go build -o main ./cmd/app && \
    chmod +x main

# Create a lightweight final image with just the executable
FROM ubuntu:latest
ENV TZ=UTC

# Install tzdata in the final image
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates tzdata wget curl && \
    rm -rf /var/lib/apt/lists/*

# Create a non-root user for security
RUN useradd -m -U -u 1001 appuser
USER appuser

# Set the working directory
WORKDIR /app

# Copy the executable from the builder stage
COPY --from=builder /app/main .

# Expose the port
EXPOSE 4000

# Use exec form for ENTRYPOINT to correctly interpret environment variable
ENTRYPOINT ["/bin/bash", "-c", "./main"]