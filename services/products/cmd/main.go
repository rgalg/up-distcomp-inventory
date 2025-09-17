package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	products_controller "products-service/internal/controller"
	products_handler_http "products-service/internal/handler"
	products_repository "products-service/internal/repository"

	"github.com/gorilla/mux"
	consul "products-service/pkg/consul"
)

func main() {
	var err error
	//var ctx context.Context

	var port int
	var datarepo *products_repository.DataRepo_Products
	var controller *products_controller.Controller_Products
	var handler *products_handler_http.Handler_Products

	// -------------------------------------------------------------------
	// variable initialization
	// -------------------------------------------------------------------
	// getting the service port from environment variable or defaulting to 8001
	port, err = strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = 8001
	}
	log.Printf("Products service starting on port %d", port)

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
	datarepo = products_repository.New()
	// controller
	controller = products_controller.New(datarepo)
	// handler
	handler = products_handler_http.New(controller)
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
	// exposing the service
	// -------------------------------------------------------------------
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), r)
	if err != nil {
		panic(err)
	}
	// -------------------------------------------------------------------
}
