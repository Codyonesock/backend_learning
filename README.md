# backend_learning

## ch1
A simple Go based app that processes a data stream and outputs statistics.
For this chapter, you'll need to configure a .env

###### .env example
```
PORT=:7000
STREAM_URL=https://stream.wikimedia.org/v2/stream/recentchange
LOG_LEVEL=INFO
JWT_SECRET=super-secure-random-key
USE_SCYLLA=TRUE
```

###### Features
- `/status`: Reads and processes recent changes from the stream.
- `/stats`: Provides aggregated statistics about the processed data.
- `/users/register`: Allows user registration.
- `/users/login`: Allows user login and returns a JWT token.

###### Example Commands
- `curl http://localhost:7000/status`
- `curl http://localhost:7000/stats` - Invalid auth attempt
- `curl -H "Authorization: Bearer <jwt-token>" http://localhost:7000/stats` - Use the token from /users/login
- `curl -X POST http://localhost:7000/users/register -H "Content-Type: application/json" -d '{"username": "blub", "password": "pw123"}'`
- `curl -X POST http://localhost:7000/users/login -H "Content-Type: application/json" -d '{"username": "blub", "password": "pw123"}'`

###### Testing
- `CGO_ENABLED=1 go test ./ch-1/internal/... -race`

###### Linting
- `golangci-lint run ./...`

## ch2
This chapter involves getting the app running in a scratch Docker container

###### Example Commands
  - `docker build -t statusapp -f ./ch-2/Dockerfile .`
  - `docker run -p 7000:7000 --env-file .env statusapp`
  - `docker run -p 7000:7000 statusapp` - Default LOG_LEVEL is INFO
  - `docker run -p 7000:7000 -e LOG_LEVEL=ERROR statusapp`

## ch3
Chapter 3 adds in docker compose, auth and a database (Scylla DB or in memory DB) to persist the stats.

###### Example Commands
- `docker compose -f ./ch-3/compose.yaml up --build` - Starts the Scylla DB and StatusApp.
- `docker compose -f ./ch-3/compose.yaml down` - Stops the DB and app.
- `docker compose -f ./ch-3/compose.yaml ps` - Verify processes are running.
- `docker compose -f ./ch-3/compose.yaml logs statusapp` - See app logs.
- `docker exec -it scylla cqlsh` - Access Scylla DB shell.
- `INTEGRATION=1 go test -tags=integration ./ch-1/internal/storage/...` - Run DB integration tests.

###### Example Stats Schema and verification
```
CREATE KEYSPACE stats_data WITH replication = {
  'class': 'SimpleStrategy',
  'replication_factor': 1
};

CREATE TABLE stats_data.stats (
  id UUID PRIMARY KEY,
  messages_consumed int,
  distinct_users map<text, int>,
  bots_count int,
  non_bots_count int,
  distinct_server_urls map<text, int>
);

Check DB:
DESCRIBE KEYSPACE stats_data;
DESCRIBE TABLE stats_data.stats;
```

## ch4
Chapter 4 adds in github actions to run unit tests, go vet, go lint, integration tests before building and uploading an image to ghcr.

## ch5
Chapter 5 introduced Redpanda and splits the app into two. A producer that reads from Wikimedia and produces to Redpanda and a consumer to read from Redpanda.

##### Example commands
- Make sure to setup your Scylla Keyspace/Table if you haven't prior. See commands in ch3.
- `docker compose -f ./ch-5/compose.yaml up --build` - Build and start all services.
- `docker compose -f ./ch-5/compose.yaml down` - Stop all services.
- `docker compose -f ./ch-5/compose.yaml ps` - Check services.
- `docker compose -f ./ch-5/compose.yaml logs` - Check logs.
- `curl http://localhost:9644/v1/status/ready` - Check Redpanda status.
- `docker exec redpanda rpk topic create wikimedia-changes` - Create Redpanda Topic

## ch6
Chapter 6 swaps the JSON to Protobuf to make the system more efficent. The producer will serialize using protobuf, and the consumer deserialize as well. The proto schema is also mounted to RedPanda console.

##### Example commands
- `protoc --go_out=. --go_opt=paths=source_relative ch-6/proto/recent_change.proto` - Generate proto code using the schema.
- `docker compose -f ./ch-6/compose.yaml up --build` - Build and start all services.

  Create Redpanda proto topic and set proto settings:
  3 partions for development. If this were a prod/larger more could be needed.
  1 Replica to save on volume. More would be needed for data that mattered to protect against failure.
- `docker exec redpanda rpk topic create wikimedia-changes-proto --partitions 3 --replicas 1`

  Set lz4 compression for low CPU overhead and efficiency:
- `docker exec redpanda rpk topic alter-config wikimedia-changes-proto --set compression.type=lz4`

## ch7
Chapter 7 brings in Prometheus & Grafana to monitor the system

##### Example commands
- Make sure your Scylla keyspaces/tables and RedPanda topics are configured. See above commands.
- `docker compose -f ./ch-7/compose.yaml up --build` - Build and start all services.
- Visit Prometheus at [http://localhost:9090](http://localhost:9090)
- Visit Grafana at [http://localhost:3000](http://localhost:3000) (default login: `admin`/`admin`)
- Producer metrics: [http://localhost:2112/metrics](http://localhost:2112/metrics)
- Consumer metrics: [http://localhost:2113/metrics](http://localhost:2113/metrics)
- Import the example dashboard from `ch-7/grafana_dashboard_example.json` in Grafana