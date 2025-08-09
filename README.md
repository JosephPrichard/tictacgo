# Tic Tac Go
A server to play tic-tac-toe written using Go, GRPC and SQLc.
the project contains a simple "component testing" solution using a temporary postgres instance.
You can interact with the server using a Postman GRPC client.

## Build

Generate Grpc Services

`protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pb/tictacgo.proto`

Generate DB Client

`$ sqlc generate`

## Execution

Create a .env file
```
DB_USER=<user>
DB_NAME=<name>
DB_PASSWORD=<password>
DB_PORT=<db-port>
SERVER_PORT=<server-port>
```

Run the server

`$ go run main.go`

## Tests

Run all tests

`$ go clean -testcache`
`$ go test ./...`

The tests run against a live database, so they take a while to run. About ~14 seconds on my machine.
<img width="494" height="125" alt="Screenshot 2025-08-08 183530" src="https://github.com/user-attachments/assets/2a4ff7bd-31b6-4006-81d3-cc517787f8b3" />
