FROM golang:1.24-alpine@sha256:7772cb5322baa875edd74705556d08f0eeca7b9c4b5367754ce3f2f00041ccee AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o enqueue-service .

# Use a small image for the final container
FROM alpine:3.21.3@sha256:a8560b36e8b8210634f77d9f7f9efd7ffa463e380b75e2e74aff4511df3ef88c

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/enqueue-service .

# Expose the service port
EXPOSE 8080

# Run the service
CMD ["./enqueue-service"]