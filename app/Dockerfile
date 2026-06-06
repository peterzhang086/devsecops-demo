# syntax=docker/dockerfile:1.7

# ── Stage 1: build ────────────────────────────────────────────────────────────
FROM golang:1.25.11-alpine AS build

# Pin via digest in week 2; for now pin by tag
WORKDIR /src

# Cache deps separately from source
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# Static binary, no cgo, strip symbols
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/server ./cmd

# ── Stage 2: runtime ──────────────────────────────────────────────────────────
# distroless static — no shell, no package manager, ~2MB base
# This will be praised by Trivy + Dockle in week 2
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/server /server

USER nonroot:nonroot
EXPOSE 8080

# OCI labels — picked up by image scanners and registries
LABEL org.opencontainers.image.title="crypto-asset-api" \
      org.opencontainers.image.source="https://github.com/peterzhang086/devsecops-demo" \
      org.opencontainers.image.licenses="MIT"

ENTRYPOINT ["/server"]
