FROM golang:1.26.0-alpine AS builder

WORKDIR /app/

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o book-service .

FROM alpine:3.23.3

WORKDIR /root/

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/book-service ./

RUN adduser -D appuser

USER appuser

EXPOSE 8080

CMD ["./book-service"]