FROM golang:1.25.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/api /app/api

EXPOSE 9000

CMD ["/app/api"]
