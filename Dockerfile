FROM golang:1.25-alpine AS base
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

FROM base AS builder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api/main.go

FROM alpine:latest AS prod
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /app/api .
EXPOSE 8080
CMD ["./api"]
