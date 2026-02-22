FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o solution-hub ./cmd/main.go


FROM alpine:latest

WORKDIR /app

RUN adduser -D appuser

COPY --from=builder /app/solution-hub .

USER appuser

EXPOSE 50051

ENTRYPOINT ["./solution-hub"]