# -------------------- Build Stage --------------------
FROM golang:1.22-alpine AS builder

# Enable Go modules
ENV GO111MODULE=on

# Create app directory
WORKDIR /app

# Install git (kadang dibutuhkan untuk go get / private modules)
RUN apk add --no-cache git

# Copy go.mod & go.sum first (biar cache build efisien)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of your source code
COPY . .

# Build binary (static binary, no CGO)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

# -------------------- Run Stage --------------------
FROM alpine:latest

# Create working directory
WORKDIR /app

# Copy binary dari stage builder
COPY --from=builder /app/server .

# Expose port aplikasi lo (ubah sesuai port server Go lo)
EXPOSE 8080

# Jalankan aplikasi
ENTRYPOINT main.go
