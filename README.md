# Tic Tac Go
A grpc server and client to play tic tac toe written using Go and sqlc.

### Build & Execution

Gemerate Grpc Services

`$ protoc --go_out=. --go_opt=paths=source_relative \             
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    service/tictacgo.proto`

Generate DB Client

`$ sqlc generate`