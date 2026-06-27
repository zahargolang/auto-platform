FROM golang:1.25.0-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o /app/cmd/api/exe /app/cmd/api

FROM debian:bookworm-slim AS runner
WORKDIR /app
COPY --from=builder /app/cmd/api/exe /app
ENV PROJECT_ROOT=/app
CMD ["/app/exe"]
