package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	products_controller "products-service/internal/controller"
	products_handler_http "products-service/internal/handler"
	products_repository "products-service/internal/repository"

	"github.com/gorilla/mux"
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
	// getting the service port from environment variable or defaulting to 8002
	port, err = strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = 8001
	}
	log.Printf("Products service starting on port %d", port)
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
	// GET all products
	r.Handle("/products", products_handler_http.AddCORSHeaders(http.HandlerFunc(handler.Get_All))).Methods(http.MethodGet)
	// GET product by productId
	r.Handle("/products/{productId}", products_handler_http.AddCORSHeaders(http.HandlerFunc(handler.Get_ByProductID))).Methods(http.MethodGet)
	// POST create product
	r.Handle("/products", products_handler_http.AddCORSHeaders(http.HandlerFunc(handler.Create_Product))).Methods(http.MethodPost)
	// CORS preflight (OPTIONS) requests for all endpoints
	r.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodOptions)
	r.HandleFunc("/products/{productId}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodOptions)
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
