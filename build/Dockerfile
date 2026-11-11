# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /echoapp .

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D -u 1000 echoapp
USER echoapp

COPY --from=builder /echoapp /usr/local/bin/echoapp
COPY --from=builder /app/migrations /app/migrations

WORKDIR /app

EXPOSE 8000

ENTRYPOINT ["echoapp"]
