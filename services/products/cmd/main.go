package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	products_controller "products-service/internal/controller"
	products_handler_http "products-service/internal/handler"
	products_repository "products-service/internal/repository"

	pb "products-service/proto/products"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

func main() {
	var err error
	//var ctx context.Context

	var port int
	var grpcPort int
	var datarepo *products_repository.DataRepo_Products
	var controller *products_controller.Controller_Products
	var handler *products_handler_http.Handler_Products
	var grpcHandler *products_handler_http.Handler_Products_GRPC

	// -------------------------------------------------------------------
	// variable initialization
	// -------------------------------------------------------------------
	// getting the service port from environment variable or defaulting to 8001
	port, err = strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = 8001
	}
	log.Printf("Products service starting on HTTP port %d", port)

	// getting the gRPC port from environment variable or defaulting to 9001
	grpcPort, err = strconv.Atoi(os.Getenv("GRPC_PORT"))
	if err != nil {
		grpcPort = 9001
	}
	log.Printf("Products service starting on gRPC port %d", grpcPort)

	// setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// initializing context
	//ctx = context.Background()
	// volatile data repository
	datarepo = products_repository.New()
	// controller
	controller = products_controller.New(datarepo)
	// handler
	handler = products_handler_http.New(controller)
	// gRPC handler
	grpcHandler = products_handler_http.NewGRPC(controller)
	// -------------------------------------------------------------------

	// -------------------------------------------------------------------
	// service endpoints
	// -------------------------------------------------------------------
	r := mux.NewRouter()
	// CORS preflight (OPTIONS) requests for all endpoints
	r.PathPrefix("/products").Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		products_handler_http.AddCORSHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(w, r)
	})
	// GET all products
	r.Handle("/products", products_handler_http.AddCORSHeaders(http.HandlerFunc(handler.Get_All))).Methods(http.MethodGet)
	// GET product by productId
	r.Handle("/products/{productId}", products_handler_http.AddCORSHeaders(http.HandlerFunc(handler.Get_ByProductID))).Methods(http.MethodGet)
	// POST create product
	r.Handle("/products", products_handler_http.AddCORSHeaders(http.HandlerFunc(handler.Create_Product))).Methods(http.MethodPost)
	// Health check endpoint
	r.Handle("/health", products_handler_http.AddCORSHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Products service is healthy")
	}))).Methods(http.MethodGet)
	// -------------------------------------------------------------------

	// -------------------------------------------------------------------
	// Start gRPC server
	// -------------------------------------------------------------------
	grpcServer := grpc.NewServer()
	pb.RegisterProductServiceServer(grpcServer, grpcHandler)

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
