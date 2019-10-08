# api

[![Go Report Card](https://goreportcard.com/badge/github.com/trandoshan-io/api)](https://goreportcard.com/report/github.com/trandoshan-io/api)

API is a Go written program used by other process to access data


## features

- use scalable database system

## how it work

- The API process connect to the database (specified by env variable *NATS_URI*)
- Start a http API using gorilla/mux and wait for request