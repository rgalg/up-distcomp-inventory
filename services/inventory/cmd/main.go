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

	"github.com/gorilla/mux"
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
	// service endpoints
	// -------------------------------------------------------------------
	r := mux.NewRouter()
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
	// CORS preflight (OPTIONS) for all endpoints
	r.HandleFunc("/inventory", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodOptions)
	r.HandleFunc("/inventory/{productId}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodOptions)
	r.HandleFunc("/inventory/{productId}/reserve", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodOptions)
	r.HandleFunc("/inventory/{productId}/release_reservation", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodOptions)
	r.HandleFunc("/inventory/{productId}/fulfill", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodOptions)
	// Health check endpoint
	r.Handle("/health", inventory_handler_http.AddCORSHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Inventory service is healthy")
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
