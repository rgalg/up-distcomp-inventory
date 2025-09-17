package products_handler_http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	products_controller "products-service/internal/controller"
	dmodel "products-service/pkg"
)

func AddCORSHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// CORS preflight request (OPTIONS) handling
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type Handler_Products struct {
	controller *products_controller.Controller_Products
}

func New(controller *products_controller.Controller_Products) *Handler_Products {
	return &Handler_Products{
		controller: controller,
	}
}

func (h *Handler_Products) Get_All(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	// getting the controller's response
	items, err := h.controller.Get_All(ctx)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// encoding the response to JSON
	err = json.NewEncoder(w).Encode(items)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler_Products) Get_ByProductID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	productId, err := strconv.Atoi(vars["productId"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// getting the controller's response
	item, err := h.controller.Get_ByProductID(ctx, productId)
	if err != nil {
		http.Error(w, "Item (product) not found", http.StatusNotFound)
		return
	}

	// encoding the response to JSON
	err = json.NewEncoder(w).Encode(item)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler_Products) Create_Product(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	var product dmodel.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// getting the controller's response
	createdItem, err := h.controller.Create_Product(ctx, &product)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	// encoding the response to JSON
	err = json.NewEncoder(w).Encode(createdItem)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
