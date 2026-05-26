# Stage 1: Builder
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
# CGO_ENABLED=0 is important for static linking, making the binary self-contained
# -o specifies the output file name
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-weather .

# Stage 2: Runner
FROM alpine:latest

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/go-weather .

# Run as the built-in unprivileged 'nobody' user (UID/GID 65534)
USER nobody:nobody

# Expose the port the application listens on
EXPOSE 8080

# Command to run the executable
CMD ["./go-weather"]
