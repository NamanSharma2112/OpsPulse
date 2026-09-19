# Build the API as a static binary, then ship it on a distroless-style base.
FROM golang:1.24-alpine AS build

WORKDIR /src

# Copy manifests first so dependency downloads cache across source changes.
COPY services/api/go.mod services/api/go.sum ./
RUN go mod download

COPY services/api/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.20

# Certificates for outbound HTTPS to the GitHub API; curl for the healthcheck.
RUN apk add --no-cache ca-certificates curl \
    && adduser -D -u 10001 opspulse

COPY --from=build /out/api /usr/local/bin/api

USER opspulse
EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 \
    CMD curl -fsS http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/usr/local/bin/api"]
