FROM golang:1.26.0-alpine AS builder

WORKDIR /app/

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o user-service .

FROM alpine:3.23.3

WORKDIR /root/

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/user-service ./

RUN adduser -D appuser

USER appuser

EXPOSE 8081

CMD ["./user-service"]