package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	"github.com/eduardovfaleiro/gatekeeper/internal/handler"
	"github.com/eduardovfaleiro/gatekeeper/internal/interceptor"
	"github.com/eduardovfaleiro/gatekeeper/internal/repository"
	"github.com/eduardovfaleiro/gatekeeper/internal/service"
	authpb "github.com/eduardovfaleiro/gatekeeper/proto"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
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
		log.Fatal("Could not connect to PostgreSQL:", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		log.Fatal("Could not connect to Redis:", err)
	}

	repo := repository.NewPostgresUserRepository(db)

	emailSvc := service.NewEmailService()

	svc := service.NewAuthService(repo, rdb, emailSvc)
	authHandler := handler.NewAuthHandler(svc)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.AuthInterceptor(os.Getenv("JWT_SECRET"))),
	)

	authpb.RegisterAuthServiceServer(server, authHandler)

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
