# Compile
FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./ch-1/cmd/statusApp

# Run
FROM scratch
COPY --from=builder /app/main /
COPY --from=golang:1.22 /etc/ssl/certs/ /etc/ssl/certs/
EXPOSE 7000
ENV PORT=:7000
ENV STREAM_URL=https://stream.wikimedia.org/v2/stream/recentchange
ENV LOG_LEVEL=INFO
ENTRYPOINT ["/main"]