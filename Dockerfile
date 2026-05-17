FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o llmbridge ./cmd/llmbridge

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/llmbridge .
COPY provider_config.yaml .
COPY templates/ ./templates/
EXPOSE 9000
CMD ["./llmbridge"]
