package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"TicTacGo/database"
	"TicTacGo/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

var port = flag.Int("port", 50051, "The server port")

func main() {
	flag.Parse()

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, "user=postgres dbname=tictacgo")
	if err != nil {
		log.Fatalf("failed to connect to database with err: %v", err)
	}
	defer pool.Close()

	server := GrpcServer{queries: database.New(pool), pool: pool}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	service.RegisterTicTacGoServiceServer(s, &server)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
