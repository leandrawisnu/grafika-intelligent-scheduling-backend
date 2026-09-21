# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS dependency-stage

WORKDIR /src
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

FROM --platform=$BUILDPLATFORM dependency-stage AS build-stage

ARG TARGETARCH
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/main ./src

FROM alpine:3.21 AS runner-stage

RUN apk add --no-cache ca-certificates curl tzdata \
    && adduser -D -H -u 10001 appuser

WORKDIR /app

COPY --from=build-stage --chown=appuser:appuser /out/main ./main
COPY --from=build-stage --chown=appuser:appuser /src/database/migrations ./database/migrations

ENV APP_HOST=0.0.0.0 \
    APP_PORT=8080 \
    TZ=Asia/Jakarta

EXPOSE 8080

USER appuser

HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -fsS "http://127.0.0.1:${APP_PORT}/health" >/dev/null || exit 1

CMD ["./main"]
