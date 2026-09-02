FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /as-ip-api ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates git

WORKDIR /app
COPY --from=builder /as-ip-api .
COPY db/init.sql ./db/init.sql

EXPOSE 3000
CMD ["./as-ip-api"]
