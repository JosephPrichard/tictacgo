package main

import (
	"flag"
)

var port = flag.Int("port", 50051, "The server port")

func main() {
	flag.Parse()
	StartGrpcServer(ServerConfig{
		port:       *port,
		dbUser:     "postgres",
		dbName:     "tictacgo",
		dbPassword: "password123",
	})
}
