# --- Builder Stage ---
# Explicitly set the user to root to ensure permissions,
# although this is the default and should not be necessary.
FROM golang:1.24-alpine AS builder
USER root

# Update packages and install git
# Combining update and add in one layer can be slightly more efficient
RUN apk update && apk add --no-cache git

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
# Using CGO_ENABLED=0 creates a static binary which is better for Alpine
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o main cmd/server/main.go

# --- Final Stage ---
FROM alpine:3.19
USER root

# Install essential certificates and timezone data
RUN apk update && apk --no-cache add ca-certificates tzdata

# Set timezone
ENV TZ=Asia/Jakarta

WORKDIR /root/

# Copy the built binary and necessary files from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/.env .
COPY --from=builder /app/templates ./templates

EXPOSE 8000

# Run the application
CMD ["./main"]
