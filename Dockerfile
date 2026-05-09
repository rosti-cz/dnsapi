# Build stage
FROM golang:1.26-alpine AS builder
# Build in the same path where source will be at runtime so Sentry stack frames
# point to the correct file locations inside the container.
RUN apk add --no-cache gcc musl-dev
WORKDIR /usr/local/src/dnsapi
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 CGO_CFLAGS="-D_LARGEFILE64_SOURCE" go build -o /dnsapi .

# Runtime stage
FROM alpine:3.22
RUN apk add --no-cache bind bind-tools sqlite-libs ca-certificates
COPY --from=builder /dnsapi /usr/local/bin/dnsapi
# Source files kept so Sentry can resolve stack frames to actual code
COPY --from=builder /usr/local/src/dnsapi /usr/local/src/dnsapi
EXPOSE 1323
ENTRYPOINT ["/usr/local/bin/dnsapi"]
