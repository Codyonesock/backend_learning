FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o consumer ./ch-1/cmd/consumer

FROM scratch
COPY --from=builder /app/consumer /consumer
COPY --from=golang:1.24 /etc/ssl/certs/ /etc/ssl/certs/
ENTRYPOINT ["/consumer"]