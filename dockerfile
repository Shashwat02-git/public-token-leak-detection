FROM golang:1.25.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /token-scanner .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /token-scanner .

COPY inventory.json .
COPY templates/ ./templates/
COPY source_files/ ./source_files/

CMD ["./token-scanner"]