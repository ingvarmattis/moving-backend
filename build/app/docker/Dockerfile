FROM golang:1.24.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o moving-service ./cmd/main.go

FROM alpine:3.21 AS runner

WORKDIR /app
COPY --from=builder /app/moving-service .

EXPOSE 8000

ENTRYPOINT ["./moving-service"]
