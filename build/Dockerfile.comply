FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /comply ./cmd/comply

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /comply /app/comply
COPY migrations /app/migrations
EXPOSE 8011
ENV COMPLY_PORT=8011
CMD ["/app/comply"]
