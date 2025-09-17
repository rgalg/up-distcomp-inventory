package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gorilla/mux"

	orders_controller "orders-service/internal/controller"
	orders_handler_http "orders-service/internal/handler"
	orders_repository "orders-service/internal/repository"
	consul "orders-service/pkg/consul"
)

func main() {
	var err error
	//var ctx context.Context

	var port int
	var datarepo *orders_repository.DataRepo_Orders
	var controller *orders_controller.Controller_Orders
	var handler *orders_handler_http.Handler_Orders

	// -------------------------------------------------------------------
	// variable initialization
	// -------------------------------------------------------------------
	// getting the service port from environment variable or defaulting to 8003
	port, err = strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = 8003
	}
	log.Printf("Orders service starting on port %d", port)

	// Initialize Consul client
	consulClient, err := consul.NewClient()
	if err != nil {
		log.Printf("Failed to create consul client: %v", err)
		log.Printf("Continuing without service discovery...")
		consulClient = nil
	} else {
		// Wait for Consul to be available
		err = consulClient.WaitForConsul(10)
		if err != nil {
			log.Printf("Consul not available: %v", err)
			log.Printf("Continuing without service discovery...")
			consulClient = nil
		} else {
			// Register service with Consul
			err = consulClient.RegisterService()
			if err != nil {
				log.Printf("Failed to register service with Consul: %v", err)
			}

			// Setup graceful shutdown to deregister service
			go func() {
				sigChan := make(chan os.Signal, 1)
				signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
				<-sigChan
				log.Println("Received shutdown signal, deregistering service...")
				if consulClient != nil {
					consulClient.DeregisterService()
				}
				os.Exit(0)
			}()
		}
	}

	// initializing context
	//ctx = context.Background()
	// volatile data repository
	datarepo = orders_repository.New()
	// controller
	controller = orders_controller.New(datarepo)
	// handler
	handler = orders_handler_http.New(controller, consulClient)
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
	// exposing the service
	// -------------------------------------------------------------------
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), r)
	if err != nil {
		panic(err)
	}
	// -------------------------------------------------------------------
}
