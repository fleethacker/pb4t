# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM node:22-bookworm-slim AS frontend
WORKDIR /build/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.26 AS backend
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/scheduler ./cmd/scheduler
RUN mkdir -p /out/data

FROM gcr.io/distroless/static-debian13:nonroot
WORKDIR /app
COPY --from=backend /out/scheduler ./scheduler
COPY --from=frontend /build/web/dist ./web/dist
COPY --from=backend --chown=65532:65532 /out/data /data
ENV LISTEN_HOST=0.0.0.0 \
    PORT=8080 \
    DATABASE_PATH=/data/scheduler.db \
    SHIP_TIMEZONE=America/New_York
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/app/scheduler"]
