FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

COPY .env .env

RUN go build -o TheLeadDestroyer .

FROM alpine:latest

WORKDIR /

COPY --from=builder /app/TheLeadDestroyer /TheLeadDestroyer

RUN chmod +x /TheLeadDestroyer

EXPOSE 8080

CMD ["/TheLeadDestroyer"]
