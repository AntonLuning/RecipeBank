FROM golang:1.24.1-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o recipe-bank-api \
    cmd/api/main.go

FROM alpine:3.20.0
WORKDIR /app

COPY --from=builder /app/recipe-bank-api /app/recipe-bank-api

ENV RP_HOST=0.0.0.0
ENV RP_PORT=9876

EXPOSE 9876

CMD ["/app/recipe-bank-api"]
