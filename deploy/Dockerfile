FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates curl

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Download Tailwind CLI and build CSS
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v4.1.8/tailwindcss-linux-x64-musl \
    && chmod +x tailwindcss-linux-x64-musl \
    && ./tailwindcss-linux-x64-musl -i static/input.css -o static/style.css --minify

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -g '' appuser

COPY --from=builder /app/server .
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates
RUN mkdir -p data && chown appuser:appuser data

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

CMD ["./server"]
