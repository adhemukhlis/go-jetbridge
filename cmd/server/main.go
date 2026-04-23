package main

import (
	"database/sql" // Provides a generic interface for SQL databases
	"fmt"          // Implements formatted I/O for console output
	"log"          // Handles logging of operational messages and errors
	"net"          // Provides network I/O primitives, including TCP listeners
	"os"           // Offers operating system functionality, including environment variables
	"os/signal"    // Provides access to incoming operating system signals
	"syscall"      // Contains low-level system call constants (e.g., SIGTERM)

	"go-jetbridge/gen/proto/role"
	"go-jetbridge/gen/proto/user" // Generated gRPC code for the User service
	pkgrole "go-jetbridge/internal/core/role"
	pkguser "go-jetbridge/internal/core/user" // Core business logic layer
	"go-jetbridge/internal/infrastructure/cache"
	"go-jetbridge/internal/middleware" // Custom middleware for API Key security
	"time"

	"buf.build/go/protovalidate"
	"github.com/joho/godotenv" // Library for loading configuration from .env files
	_ "github.com/lib/pq"      // PostgreSQL driver (imported for side effects)
	"google.golang.org/grpc"   // Core gRPC framework and utilities
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load environment variables from .env file for local configuration
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	dbPort := os.Getenv("PGPORT")
	if dbPort == "" {
		log.Fatal("PGPORT is not set in the environment")
	}

	dbUser := os.Getenv("PGUSER")
	if dbUser == "" {
		log.Fatal("PGUSER is not set in the environment")
	}

	dbPassword := os.Getenv("PGPASSWORD")

	dbHost := os.Getenv("PGHOST")
	if dbHost == "" {
		log.Fatal("PGHOST is not set in the environment")
	}

	dbName := os.Getenv("PGDATABASE")
	if dbName == "" {
		log.Fatal("PGDATABASE is not set in the environment")
	}

	sslMode := os.Getenv("SSL_MODE")
	if sslMode == "" {
		log.Fatal("SSL_MODE is not set in the environment")
	}

	// Retrieve database connection string from environment
	dbURL := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s", dbUser, dbPassword, dbHost, dbPort, dbName, sslMode)

	// Initialize SQL database connection pool using the postgres driver
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	// Configure connection pool parameters for resource management
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	// Verify database connectivity before proceeding with server startup
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to verify database connection: %v", err)
	}

	// Configure the network listener for the gRPC server
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to establishing network listener: %v", err)
	}

	// Initialize protovalidate validator
	v, err := protovalidate.New()
	if err != nil {
		log.Fatalf("failed to initialize validator: %v", err)
	}

	// Initialize gRPC server with a chain of Unary Interceptors for security, validation and error handling
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.UnaryErrorInterceptor,
			middleware.UnaryAuthInterceptor,
			middleware.UnaryValidationInterceptor(v),
		),
	)

	// Dependency Injection: Initialize standard utilities, repository, service, and transport handlers

	// Initialize In-Memory Cache (Default TTL: 5m, Cleanup: 10m)
	appCache := cache.NewInMemoryCache(5*time.Minute, 10*time.Minute)

	userRepo := &pkguser.User{DB: db}
	userService := pkguser.NewService(userRepo, appCache)
	userHandler := &pkguser.Handler{Service: userService}
	user.RegisterUserServiceServer(s, userHandler)

	roleRepo := &pkgrole.Role{DB: db}
	roleService := pkgrole.NewService(roleRepo, appCache)
	roleHandler := &pkgrole.Handler{Service: roleService}
	role.RegisterRoleServiceServer(s, roleHandler)

	// Register reflection service on gRPC server to support Postman/grpcurl discovery
	reflection.Register(s)

	// Launch gRPC server in a background goroutine to allow for non-blocking signal handling
	go func() {
		fmt.Printf("🚀 gRPC Bridge Server running on :%s\n", port)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to start gRPC service: %v", err)
		}
	}()

	// Block main execution and wait for OS termination signals (SIGINT, SIGTERM)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Initiate graceful shutdown procedure to allow active RPCs to complete
	log.Println("Initiating graceful shutdown...")
	s.GracefulStop()

	// Release database resources
	db.Close()
	log.Println("Server successfully stopped.")
}
