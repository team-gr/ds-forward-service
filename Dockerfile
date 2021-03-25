FROM golang:1.15-alpine as builder
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY src/go.mod go.mod
COPY src/go.sum go.sum
RUN go mod download
COPY src .
RUN CGO_ENABLED=0 GOOS=linux go build -o forwarder.bin /app/main.go

FROM scratch
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/forwarder.bin /app/forwarder.bin