# Tic Tac Go
A server to play tic-tac-toe written using Go, GRPC and SQLc.
the project contains a simple "component testing" solution using a temporary postgres instance.
You can interact with the server using a Postman GRPC client.

<img width="1417" height="884" alt="image" src="https://github.com/user-attachments/assets/b85d1a70-caff-4c06-a543-a4a473091478" />

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

<img width="494" height="125" alt="image" src="https://github.com/user-attachments/assets/72673ef6-696d-4af3-bb03-3d2aaa6bf0a6" />