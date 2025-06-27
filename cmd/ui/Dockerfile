FROM golang:1.24.1-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o recipe-bank-ui \
    cmd/ui/main.go

FROM alpine:3.20.0
WORKDIR /app

COPY --from=builder /app/recipe-bank-ui /app/recipe-bank-ui
COPY bin/assets /app/assets

ENV RP_UI_DEBUG=false
ENV RP_UI_HOST=0.0.0.0
ENV RP_UI_PORT=9999
ENV RP_UI_ASSETS_PATH=/app/assets

EXPOSE 9999

CMD ["/app/recipe-bank-ui"]
