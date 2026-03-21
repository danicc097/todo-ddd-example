FROM golang:1.25-alpine AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

FROM base AS builder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/api ./cmd/api/main.go

FROM alpine:3.21 AS prod
RUN apk add --no-cache ca-certificates && \
    addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=builder /app/api .
USER app
EXPOSE 8080
CMD ["./api"]
