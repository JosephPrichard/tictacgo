# Tic Tac Go
A grpc server to play tic-tac-toe written using Go and sqlc.

### Build & Execution

Generate Grpc Services

`protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pb/tictacgo.proto`

Generate DB Client

`$ sqlc generate`

Execute Server

`$ go run main.go`