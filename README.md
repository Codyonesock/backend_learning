# backend_learning

## ch1
A simple Go based app that processes a data stream and outputs statistics.
For this chapter, you'll need to configure a .env

###### .env example
export PORT=:7000
export STREAM_URL=https://stream.wikimedia.org/v2/stream/recentchange

###### Features
- `/status`: Reads and processes recent changes from the stream.
- `/stats`: Provides aggregated statistics about the processed data.

###### Example Commands
- `curl http://localhost:7000/status`
- `curl http://localhost:7000/stats`

###### Testing
- `CGO_ENABLED=1 go test ./ch-1/internal/... -race`

## ch2
This chapter involves getting the app running in a scratch Docker container

###### Example Commands
  - `docker build -t statusapp -f ./ch-2/Dockerfile .`
  - `docker run -p 7000:7000 statusapp`

## ch3
Chapter 3 will involve bringing in a Scylla DB to persist the stats.