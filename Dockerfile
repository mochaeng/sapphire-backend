# syntax=docker/dockerfile:1

# build stage
ARG GO_VERSION=1.23.0
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS builder

# set the Go module path
# ENV GO111MODULE=on

WORKDIR /build

# Copy the entire project
COPY . .

# download dependencies and build
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o sapphire ./cmd/sapphire/main.go

# final stage
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates

# create non-root user
RUN adduser \
    -D \
    -g "" \
    -u 10001 \
    appuser

# create necessary directories with correct permissions
RUN mkdir -p /app/data /app/migrate/migrations && \
    chown -R appuser:appuser /app

USER appuser

# copy binary and migrations
COPY --from=builder --chown=appuser:appuser /build/sapphire /app/
COPY --from=builder --chown=appuser:appuser /build/migrate/migrations /app/migrate/migrations

EXPOSE 7777

HEALTHCHECK --interval=30s --timeout=3s \
    CMD wget --spider http://localhost:7777/health || exit 1

CMD ["/app/sapphire"]
