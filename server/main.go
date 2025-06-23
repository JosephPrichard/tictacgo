package main

import (
	"TicTacGo/db"
	"TicTacGo/pb"
	"context"
	"flag"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"log"
	"net"
)

var port = flag.Int("port", 50051, "The serve port")

func main() {
	flag.Parse()

	dbUser := "postgres"
	dbName := "tictacgo"
	dbPassword := "password123"
	dbPort := 5432

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, fmt.Sprintf("user=%s dbname=%s password=%s port=%d", dbUser, dbName, dbPassword, dbPort))
	if err != nil {
		log.Fatalf("failed to connect to database with err: %v", err)
	}
	defer pool.Close()

	server := &GrpcServer{queries: db.New(pool), pool: pool}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	baseServer := grpc.NewServer()
	pb.RegisterTicTacGoServiceServer(baseServer, server)

	log.Printf("serve listening at %v", lis.Addr())
	if err := baseServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
