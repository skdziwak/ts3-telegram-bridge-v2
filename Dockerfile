# Start from the official Golang base image
FROM golang:1.18 as builder

# Set the working directory outside of GOPATH to enable the support for modules.
WORKDIR /src

# Copy go.mod and go.sum to download all dependencies
COPY go.* ./
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bridge .

# Start a new stage from scratch
FROM alpine:latest

# Install ca-certificates in case the application makes outgoing HTTPS requests
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /app

# Copy the pre-built binary file from the previous stage
COPY --from=builder /src/bridge .

# Copy the config file
COPY config.yaml .

# Command to run the executable
CMD ["./bridge"]
