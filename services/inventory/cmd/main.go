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

	inventory_controller "inventory-service/internal/controller"
	inventory_handler_http "inventory-service/internal/handler"
	inventory_repository "inventory-service/internal/repository"

	pb "inventory-service/proto/inventory"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

func main() {
	var err error
	//var ctx context.Context

	var port int
	var grpcPort int
	var datarepo *inventory_repository.DataRepo_Inventory
	var controller *inventory_controller.Controller_Inventory
	var handler *inventory_handler_http.Handler_Inventory
	var grpcHandler *inventory_handler_http.Handler_Inventory_GRPC

	// -------------------------------------------------------------------
	// variable initialization
	// -------------------------------------------------------------------
	// getting the service port from environment variable or defaulting to 8002
	port, err = strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = 8002
	}
	log.Printf("Inventory service starting on HTTP port %d", port)

	// getting the gRPC port from environment variable or defaulting to 9002
	grpcPort, err = strconv.Atoi(os.Getenv("GRPC_PORT"))
	if err != nil {
		grpcPort = 9002
	}
	log.Printf("Inventory service starting on gRPC port %d", grpcPort)

	// setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// initializing context
	//ctx = context.Background()
	// volatile data repository
	datarepo = inventory_repository.New()
	// controller
	controller = inventory_controller.New(datarepo)
	// handler
	handler = inventory_handler_http.New(controller)
	// gRPC handler
	grpcHandler = inventory_handler_http.NewGRPC(controller)
	// -------------------------------------------------------------------

	// -------------------------------------------------------------------
	// service endpoints
	// -------------------------------------------------------------------
	r := mux.NewRouter()
	// CORS preflight (OPTIONS) requests for all endpoints
	r.PathPrefix("/inventory").Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inventory_handler_http.AddCORSHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(w, r)
	})
	// GET all inventory
	r.Handle("/inventory", inventory_handler_http.AddCORSHeaders(http.HandlerFunc(handler.Get_All))).Methods(http.MethodGet)
	// GET inventory by productId
	r.Handle("/inventory/{productId}", inventory_handler_http.AddCORSHeaders(http.HandlerFunc(handler.Get_ByProductID))).Methods(http.MethodGet)
	// PUT update stock
	r.Handle("/inventory/{productId}", inventory_handler_http.AddCORSHeaders(http.HandlerFunc(handler.Update_Stock))).Methods(http.MethodPut)
	// POST reserve stock
	r.Handle("/inventory/{productId}/reserve", inventory_handler_http.AddCORSHeaders(http.HandlerFunc(handler.Reserve_Stock))).Methods(http.MethodPost)
	// POST release reservation
	r.Handle("/inventory/{productId}/release_reservation", inventory_handler_http.AddCORSHeaders(http.HandlerFunc(handler.Release_Reservation))).Methods(http.MethodPost)
	// POST fulfill reservation
	r.Handle("/inventory/{productId}/fulfill", inventory_handler_http.AddCORSHeaders(http.HandlerFunc(handler.Fulfill_Reservation))).Methods(http.MethodPost)
	// Health check endpoint
	r.Handle("/health", inventory_handler_http.AddCORSHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Inventory service is healthy")
	}))).Methods(http.MethodGet)
	// -------------------------------------------------------------------

	// -------------------------------------------------------------------
	// Start gRPC server
	// -------------------------------------------------------------------
	grpcServer := grpc.NewServer()
	pb.RegisterInventoryServiceServer(grpcServer, grpcHandler)

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
