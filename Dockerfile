# stage 1: build
FROM golang:1.23-alpine AS build

WORKDIR /app

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the binary
RUN go build -o server .

# stage 2: runtime
FROM alpine:3.19

WORKDIR /root/
COPY --from=build /app/server .

EXPOSE 3000
CMD ["./server"]
