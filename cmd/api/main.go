package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/eduardovfaleiro/gatekeeper/internal/handler"
	"github.com/eduardovfaleiro/gatekeeper/internal/interceptor"
	"github.com/eduardovfaleiro/gatekeeper/internal/repository"
	"github.com/eduardovfaleiro/gatekeeper/internal/service"
	"github.com/eduardovfaleiro/gatekeeper/internal/worker"
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

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	db, err := sql.Open("postgres", dsn)

	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Could not connect to PostgreSQL:", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	err = rdb.XGroupCreateMkStream(ctx, "email_stream", "email_processors", "0").Err()
	if err != nil {
		log.Printf("Aviso: Stream ou Grupo de Redis já existem ou erro: %v", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		log.Fatal("Could not connect to Redis:", err)
	}

	repo := repository.NewPostgresUserRepository(db)

	emailSvc := service.NewEmailService(os.Getenv("SMTP_HOST"),
		os.Getenv("SMTP_PORT"),
		os.Getenv("SMTP_USER"),
		os.Getenv("SMTP_PASS"),
		os.Getenv("SMTP_FROM"))
	emailWorker := worker.NewEmailWorker(rdb, emailSvc)

	go emailWorker.Start(ctx)

	svc := service.NewAuthService(repo, rdb, emailSvc)
	authHandler := handler.NewAuthHandler(svc)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.AuthInterceptor(os.Getenv("JWT_SECRET"))),
	)

	authpb.RegisterAuthServiceServer(grpcServer, authHandler)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("gRPC server running on :50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
