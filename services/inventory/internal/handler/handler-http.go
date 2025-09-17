package inventory_handler_http

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	inventory_controller "inventory-service/internal/controller"
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

type Handler_Inventory struct {
	controller *inventory_controller.Controller_Inventory
}

func New(controller *inventory_controller.Controller_Inventory) *Handler_Inventory {
	return &Handler_Inventory{
		controller: controller,
	}
}

func (h *Handler_Inventory) Get_All(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	// getting the controller's response
	items, err := h.controller.Get_All(ctx)
	if err != nil {
		log.Printf("Error getting all inventory items: Repository error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// encoding the response to JSON
	err = json.NewEncoder(w).Encode(items)
	if err != nil {
		log.Printf("Error encoding inventory items to JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler_Inventory) Get_ByProductID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	r_params := mux.Vars(r)
	productID, err := strconv.Atoi(r_params["productId"])
	if err != nil {
		log.Printf("Error getting product ID from URL: %v", err)
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// getting the controller's response
	item, err := h.controller.Get_ByProductID(ctx, productID)
	if err != nil {
		log.Printf("Error getting inventory item by product ID: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// encoding the response to JSON
	err = json.NewEncoder(w).Encode(item)
	if err != nil {
		log.Printf("Error encoding inventory item to JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler_Inventory) Update_Stock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	r_params := mux.Vars(r)
	productID, err := strconv.Atoi(r_params["productId"])
	if err != nil {
		log.Printf("Error getting product ID from URL: %v", err)
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var template_req struct {
		Stock int `json:"stock"`
	}

	if err := json.NewDecoder(r.Body).Decode(&template_req); err != nil {
		log.Printf("Error decoding JSON request body: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.controller.Update_Stock(ctx, productID, template_req.Stock); err != nil {
		log.Printf("Error updating inventory stock: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	item, err := h.controller.Get_ByProductID(ctx, productID)
	if err != nil {
		log.Printf("Error getting updated inventory item by product ID: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// encoding the response to JSON
	err = json.NewEncoder(w).Encode(item)
	if err != nil {
		log.Printf("Error encoding updated inventory item to JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	// logging
	log.Printf("Updated inventory item: %+v", item)
}

func (h *Handler_Inventory) Reserve_Stock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	r_params := mux.Vars(r)
	productID, err := strconv.Atoi(r_params["productId"])
	if err != nil {
		log.Printf("Error getting product ID from URL: %v", err)
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var template_req struct {
		Stock int `json:"stock"`
	}

	if err := json.NewDecoder(r.Body).Decode(&template_req); err != nil {
		log.Printf("Error decoding JSON request body: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.controller.Reserve_Stock(ctx, productID, template_req.Stock); err != nil {
		log.Printf("Error reserving inventory stock: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	item, err := h.controller.Get_ByProductID(ctx, productID)
	if err != nil {
		log.Printf("Error getting updated inventory item by product ID: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// encoding the response to JSON
	err = json.NewEncoder(w).Encode(item)
	if err != nil {
		log.Printf("Error encoding updated inventory item to JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	// logging
	log.Printf("Reserved inventory item: %+v", item)
}

func (h *Handler_Inventory) Release_Reservation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	r_params := mux.Vars(r)
	productID, err := strconv.Atoi(r_params["productId"])
	if err != nil {
		log.Printf("Error getting product ID from URL: %v", err)
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var template_req struct {
		Stock int `json:"stock"`
	}

	if err := json.NewDecoder(r.Body).Decode(&template_req); err != nil {
		log.Printf("Error decoding JSON request body: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.controller.Release_Reservation(ctx, productID, template_req.Stock); err != nil {
		log.Printf("Error releasing inventory reservation: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	item, err := h.controller.Get_ByProductID(ctx, productID)
	if err != nil {
		log.Printf("Error getting updated inventory item by product ID: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// encoding the response to JSON
	err = json.NewEncoder(w).Encode(item)
	if err != nil {
		log.Printf("Error encoding updated inventory item to JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	// logging
	log.Printf("Released reservation for inventory item: %+v", item)
}

func (h *Handler_Inventory) Fulfill_Reservation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	r_params := mux.Vars(r)
	productID, err := strconv.Atoi(r_params["productId"])
	if err != nil {
		log.Printf("Error getting product ID from URL: %v", err)
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var template_req struct {
		Stock int `json:"stock"`
	}

	if err := json.NewDecoder(r.Body).Decode(&template_req); err != nil {
		log.Printf("Error decoding JSON request body: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.controller.Fulfill_Reservation(ctx, productID, template_req.Stock); err != nil {
		log.Printf("Error fulfilling inventory reservation: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	item, err := h.controller.Get_ByProductID(ctx, productID)
	if err != nil {
		log.Printf("Error getting updated inventory item by product ID: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// encoding the response to JSON
	err = json.NewEncoder(w).Encode(item)
	if err != nil {
		log.Printf("Error encoding updated inventory item to JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	// logging
	log.Printf("Fulfilled reservation for inventory item: %+v", item)
}
