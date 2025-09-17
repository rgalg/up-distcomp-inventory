package inventory_handler_http

import (
	"encoding/json"
	"log"
	"net/http"

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
