# Build stage
FROM node:22-alpine AS ui-builder
WORKDIR /app/web
COPY web/package.json web/pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm build

FROM golang:1.24-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui-builder /app/web/dist ./internal/broker/ui/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o yakusoku-broker ./cmd/yakusoku-broker

# Runtime stage
FROM alpine:3.21
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=go-builder /app/yakusoku-broker .
EXPOSE 8080
ENTRYPOINT ["./yakusoku-broker"]
