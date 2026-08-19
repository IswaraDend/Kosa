# syntax=docker/dockerfile:1

# ---------- build ----------
FROM golang:1.25-alpine AS build
WORKDIR /src

# Dependencies are copied and downloaded on their own so that editing a handler
# does not invalidate the module cache and re-download the whole graph.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO is off deliberately. go.sum still lists go-sqlite3 (which would need a C
# toolchain), but nothing in this codebase imports it — the only driver actually
# used is gorm.io/driver/postgres, which is pure Go via pgx. That yields a
# static binary, so the runtime stage needs no Go toolchain and no libc.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/kosa .

# ---------- runtime ----------
FROM alpine:3.20

# ca-certificates for outbound TLS, tzdata because the DSN sets TimeZone.
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 kosa

COPY --from=build /out/kosa /usr/local/bin/kosa

USER kosa
ENV GIN_MODE=release
EXPOSE 8080

# /ping needs neither auth nor a database, so it answers as soon as the HTTP
# server is up — which is exactly what "is this container ready" should mean.
HEALTHCHECK --interval=15s --timeout=3s --start-period=20s --retries=5 \
    CMD wget -qO- http://127.0.0.1:8080/ping >/dev/null 2>&1 || exit 1

ENTRYPOINT ["kosa"]
