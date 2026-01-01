package main

import (
	"database/sql"
	"log"
	"net"

	"github.com/eduardovfaleiro/gatekeeper/internal/handler"
	"github.com/eduardovfaleiro/gatekeeper/internal/repository"
	"github.com/eduardovfaleiro/gatekeeper/internal/service"
	authpb "github.com/eduardovfaleiro/gatekeeper/proto"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db, err := sql.Open("postgres", "postgres://user:pass@localhost:5435/gatekeeper_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	repo := repository.NewPostgresUserRepository(db)
	svc := service.NewAuthService(repo)
	authHandler := handler.NewAuthHandler(svc)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	authpb.RegisterAuthServiceServer(grpcServer, authHandler)
	reflection.Register(grpcServer)

	log.Println("gRPC server running on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
