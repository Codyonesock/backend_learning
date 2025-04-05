# Compile
FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./ch-1/cmd/statusApp

# Run
FROM scratch
COPY --from=builder /app/main /
COPY --from=builder /app/config.json /
COPY --from=golang:1.22 /etc/ssl/certs/ /etc/ssl/certs/
EXPOSE 7000
ENTRYPOINT ["/main"]