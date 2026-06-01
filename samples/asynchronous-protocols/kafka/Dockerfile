# Build context is this kafka/ folder, where go.mod and go.sum live.

FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG BINARY=consumer
RUN CGO_ENABLED=0 GOOS=linux go build -o /out ./cmd/${BINARY}

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /out /app
ENTRYPOINT ["/app"]
