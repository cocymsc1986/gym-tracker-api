# syntax=docker/dockerfile:1

# Local dev / QA image for gym-tracker-api.
# Runs the API with DEV_MODE=true: in-memory repositories, stubbed auth,
# no AWS dependencies. Not used for production deploys.

FROM golang:1.21-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api

FROM alpine:3.19
RUN apk add --no-cache ca-certificates wget
WORKDIR /app
COPY --from=builder /out/api /app/api

ENV DEV_MODE=true \
    PORT=8080 \
    CORS_ALLOWED_ORIGINS=*

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=2s --start-period=2s --retries=3 \
  CMD wget -q -O - http://localhost:${PORT}/exercises/dev-user >/dev/null || exit 1

ENTRYPOINT ["/app/api"]
