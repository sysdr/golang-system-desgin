FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /app/cmd/server
RUN go build -o /usr/local/bin/server .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /usr/local/bin/server /usr/local/bin/server
EXPOSE 8080
CMD ["server"]
