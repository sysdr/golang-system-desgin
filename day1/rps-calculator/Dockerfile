# Use the official Go image to build the application
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod (go.sum may not exist for stdlib-only modules)
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o rps-calculator main.go

# Use a minimal image for the final executable
FROM alpine:latest

WORKDIR /app

# Copy the built executable from the builder stage
COPY --from=builder /app/rps-calculator .

# Command to run the executable
CMD ["./rps-calculator"]
