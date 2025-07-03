package main

import (
	"TicTacGo/db"
	"TicTacGo/pb"
	"TicTacGo/server"
	"TicTacGo/utils"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	f, err := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatalf("failed to open log file: %v", err)
		}
	}(f)
	log.SetOutput(io.MultiWriter(f, os.Stdout))

	config := utils.NewConfig()

	dbUser := config.Get("DB_USER")
	dbName := config.Get("DB_NAME")
	dbPassword := config.Get("DB_PASSWORD")
	dbPort := config.Get("DB_PORT")
	serverPort := config.Get("SERVER_PORT")

	ctx := context.Background()

	connString := fmt.Sprintf("user=%s dbname=%s password=%s port=%s", dbUser, dbName, dbPassword, dbPort)
	log.Printf("creating database connection on with connString: %s", connString)

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("failed to connect to database with err: %v", err)
	}
	defer pool.Close()

	log.Printf("starting server on port: %s", serverPort)

	serve := &server.GrpcServer{Queries: db.New(pool), Pool: pool}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", serverPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	baseServer := grpc.NewServer()
	pb.RegisterTicTacGoServiceServer(baseServer, serve)

	log.Printf("serve listening at %v", lis.Addr())
	if err := baseServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
