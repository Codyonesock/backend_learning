# backend_learning

## ch1
A simple Go based app that processes a data stream and outputs statistics.

###### Features
- `/status`: Reads and processes recent changes from the stream.
- `/stats`: Provides aggregated statistics about the processed data.

###### CMD
- `curl http://localhost:7000/status`
- `curl http://localhost:7000/stats`

## ch2
This chapter involves getting the app running in a scratch Docker container

###### CMD
  - `docker build -t statusapp -f .\ch-2\Dockerfile .`
  - `docker run -p 7000:7000 statusapp`
