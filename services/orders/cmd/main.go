package main

import (
	"database/sql"

	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	orders_controller "orders-service/internal/controller"
	orders_handler_http "orders-service/internal/handler"
	orders_repository "orders-service/internal/repository"

	pb "orders-service/proto/orders"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
	"google.golang.org/grpc"
)

func initDB() (*sql.DB, error) {
	// load .env file if it exists
	if err := godotenv.Load("../../../.env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// look for database credentials in the environment variables
	host := getEnv("DB_HOST", "")
	port := getEnv("DB_PORT", "")
	user := getEnv("DB_USER", "")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "inventory_db")
	// throw and error if any required env variable is missing
	if password == "" || host == "" || port == "" || user == "" || dbname == "" {
		return nil, fmt.Errorf("An environment variable is missing: DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME are required")
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// test the connection to the db
	if err = db.Ping(); err != nil {
		return nil, err
	}

	// set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Printf("Successfully connected to PostgreSQL at %s:%s", host, port)
	return db, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	var err error
	//var ctx context.Context

	var port int
	var grpcPort int
	var datarepo *orders_repository.DataRepo_Orders
	var controller *orders_controller.Controller_Orders
	var handler *orders_handler_http.Handler_Orders
	var grpcHandler *orders_handler_http.Handler_Orders_GRPC

	// -------------------------------------------------------------------
	// variable initialization
	// -------------------------------------------------------------------

	// initializing database connection
	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// getting the service port from environment variable or defaulting to 8003
	port, err = strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = 8003
	}
	log.Printf("Orders service starting on HTTP port %d", port)

	// getting the gRPC port from environment variable or defaulting to 9003
	grpcPort, err = strconv.Atoi(os.Getenv("GRPC_PORT"))
	if err != nil {
		grpcPort = 9003
	}
	log.Printf("Orders service starting on gRPC port %d", grpcPort)

	// Get service addresses for gRPC clients (using Kubernetes service discovery)
	productsAddr := os.Getenv("PRODUCTS_GRPC_ADDR")
	if productsAddr == "" {
		productsAddr = "products-service:9001"
	}
	inventoryAddr := os.Getenv("INVENTORY_GRPC_ADDR")
	if inventoryAddr == "" {
		inventoryAddr = "inventory-service:9002"
	}

	// setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// initializing context
	//ctx = context.Background()
	// volatile data repository
	datarepo = orders_repository.New(db)
	// controller
	controller = orders_controller.New(datarepo)
	// handler (HTTP still uses consul for backward compatibility, but pass nil now)
	handler = orders_handler_http.New(controller, nil)
	// gRPC handler
	grpcHandler, err = orders_handler_http.NewGRPC(controller, productsAddr, inventoryAddr)
	if err != nil {
		log.Fatalf("Failed to create gRPC handler: %v", err)
	}
	// -------------------------------------------------------------------

	// -------------------------------------------------------------------
	// service endpoints
	// -------------------------------------------------------------------
	r := mux.NewRouter()
	// CORS preflight (OPTIONS) requests for all endpoints
	r.PathPrefix("/orders").Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orders_handler_http.AddCORSHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(w, r)
	})
	// GET all orders
	r.Handle("/orders", orders_handler_http.AddCORSHeaders(http.HandlerFunc(handler.Get_All))).Methods(http.MethodGet)
	// GET order by ID
	r.Handle("/orders/{orderId}", orders_handler_http.AddCORSHeaders(http.HandlerFunc(handler.Get_ByOrderID))).Methods(http.MethodGet)
	// POST create order
	r.Handle("/orders", orders_handler_http.AddCORSHeaders(http.HandlerFunc(handler.Create_Order))).Methods(http.MethodPost)
	// POST fulfill order
	r.Handle("/orders/{orderId}/fulfill", orders_handler_http.AddCORSHeaders(http.HandlerFunc(handler.Fulfill_Order))).Methods(http.MethodPost)
	// -------------------------------------------------------------------
	// Health check endpoint
	r.Handle("/health", orders_handler_http.AddCORSHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Orders service is healthy")
	}))).Methods(http.MethodGet)
	// -------------------------------------------------------------------

	// -------------------------------------------------------------------
	// Start gRPC server
	// -------------------------------------------------------------------
	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, grpcHandler)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port: %v", err)
	}

	go func() {
		log.Printf("gRPC server listening on port %d", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()
	// -------------------------------------------------------------------

	// -------------------------------------------------------------------
	// Start HTTP server
	// -------------------------------------------------------------------
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	go func() {
		log.Printf("HTTP server listening on port %d", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to serve HTTP: %v", err)
		}
	}()
	// -------------------------------------------------------------------

	// -------------------------------------------------------------------
	// Wait for shutdown signal
	// -------------------------------------------------------------------
	<-sigChan
	log.Println("Received shutdown signal, shutting down gracefully...")
	grpcServer.GracefulStop()
	log.Println("Servers stopped")
	// -------------------------------------------------------------------
}
