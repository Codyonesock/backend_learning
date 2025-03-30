#* 1. Create a Dockerfile for your application
#! 2. Build & run your dockerized application
#! 3. Build a scratch container image of your application
#! 4. Use a file to set all the configurable items like ports, URLs and anything else that can be dynamic, load these configs via the file

# Build the app
FROM golang:1.22 AS builder
WORKDIR /app
COPY ch-1/cmd/statusApp/ .
RUN go build -o app .

# Setup the container and run the app
FROM scratch
WORKDIR /
COPY --from=builder /app/app .
COPY config.json /config.json
EXPOSE 7000
ENTRYPOINT ["/app"]