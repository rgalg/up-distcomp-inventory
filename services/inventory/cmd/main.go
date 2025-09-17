package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	inventory_controller "inventory-service/internal/controller"
	inventory_handler_http "inventory-service/internal/handler"
	inventory_repository "inventory-service/internal/repository"
)

func main() {
	var err error
	//var ctx context.Context

	var port int
	var datarepo *inventory_repository.DataRepo_Inventory
	var controller *inventory_controller.Controller_Inventory
	var handler *inventory_handler_http.Handler_Inventory

	// -------------------------------------------------------------------
	// variable initialization
	// -------------------------------------------------------------------
	// getting the service port from environment variable or defaulting to 8002
	port, err = strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = 8002
	}
	log.Printf("Inventory service starting on port %d", port)
	// initializing context
	//ctx = context.Background()
	// volatile data repository
	datarepo = inventory_repository.New()
	// controller
	controller = inventory_controller.New(datarepo)
	// handler
	handler = inventory_handler_http.New(controller)
	// -------------------------------------------------------------------

	// -------------------------------------------------------------------
	// router (handler)
	// -------------------------------------------------------------------
	// r := mux.NewRouter()
	// r.Use(inventory_handler_http.corsHandler)
	// -------------------------------------------------------------------

	// -------------------------------------------------------------------
	// service endpoints
	// -------------------------------------------------------------------
	// GET requests
	// r.HandleFunc("/inventory", inventory_handler_http.Handler_Inventory.getAllInventory).Methods(http.MethodGet)
	// // r.HandleFunc("/inventory/{productId}", getInventoryByProduct).Methods(http.MethodGet)
	// // // PUT requests
	// // r.HandleFunc("/inventory/{productId}", updateInventory).Methods(http.MethodPut)
	// // // POST requests
	// // r.HandleFunc("/inventory/{productId}/reserve", reserveInventory).Methods(http.MethodPost)
	// // r.HandleFunc("/inventory/{productId}/fulfill", fulfillReservation).Methods(http.MethodPost)
	// // -------------------------------------------------------------------
	// // CORS preflight (OPTIONS) requests for all endpoints
	// r.HandleFunc("/inventory", func(w http.ResponseWriter, r *http.Request) {
	// 	w.WriteHeader(http.StatusOK)
	// }).Methods(http.MethodOptions)
	// r.HandleFunc("/inventory/{productId}", func(w http.ResponseWriter, r *http.Request) {
	// 	w.WriteHeader(http.StatusOK)
	// }).Methods(http.MethodOptions)
	// r.HandleFunc("/inventory/{productId}/reserve", func(w http.ResponseWriter, r *http.Request) {
	// 	w.WriteHeader(http.StatusOK)
	// }).Methods(http.MethodOptions)
	// r.HandleFunc("/inventory/{productId}/fulfill", func(w http.ResponseWriter, r *http.Request) {
	// 	w.WriteHeader(http.StatusOK)
	// }).Methods(http.MethodOptions)
	// // -------------------------------------------------------------------
	// // health check endpoint
	// r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
	// 	w.WriteHeader(http.StatusOK)
	// 	fmt.Fprint(w, "Inventory service is healthy")
	// }).Methods(http.MethodGet)
	// -------------------------------------------------------------------

	http.Handle("/inventory", inventory_handler_http.AddCORSHeaders(http.HandlerFunc(handler.Get_All)))

	// -------------------------------------------------------------------
	// exposing the service
	// -------------------------------------------------------------------
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic(err)
	}
	// log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r)) // log and os.Exit(1)
	// -------------------------------------------------------------------
}
