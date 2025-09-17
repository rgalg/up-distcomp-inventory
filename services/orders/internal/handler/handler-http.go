package orders_handler_http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"orders-service/internal"
	orders_controller "orders-service/internal/controller"
	dmodel "orders-service/pkg"
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

type Handler_Orders struct {
	controller *orders_controller.Controller_Orders
}

func New(controller *orders_controller.Controller_Orders) *Handler_Orders {
	return &Handler_Orders{
		controller: controller,
	}
}

func (h *Handler_Orders) Get_All(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	// getting the controller's response
	orders, err := h.controller.Get_All(ctx)
	if err != nil {
		log.Printf("Error getting all orders: Repository error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// encoding the response to JSON
	err = json.NewEncoder(w).Encode(orders)
	if err != nil {
		log.Printf("Error encoding orders to JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler_Orders) Get_ByOrderID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	r_params := mux.Vars(r)
	orderID, err := strconv.Atoi(r_params["id"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	// getting the controller's response
	order, err := h.controller.Get_ByOrderID(ctx, orderID)
	if err != nil {
		if err == internal.ErrItemNotFound {
			http.Error(w, "Order not found", http.StatusNotFound)
		} else {
			log.Printf("Error getting order by ID: Repository error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	// encoding the response to JSON
	err = json.NewEncoder(w).Encode(order)
	if err != nil {
		log.Printf("Error encoding order to JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler_Orders) Create_Order(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	var template_req struct {
		CustomerID int                `json:"customer_id"`
		Items      []dmodel.OrderItem `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&template_req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(template_req.Items) == 0 {
		http.Error(w, "Order must contain at least one item", http.StatusBadRequest)
		return
	}

	// Calculate total amount and validate products
	var totalAmount float64
	for _, item := range template_req.Items {
		product, err := getProduct(item.ProductID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Product %d not found", item.ProductID), http.StatusBadRequest)
			return
		}

		totalAmount += product.Price * float64(item.Quantity)

		// Try to reserve inventory
		if err := reserveInventory(item.ProductID, item.Quantity); err != nil {
			http.Error(w, fmt.Sprintf("Failed to reserve inventory for product %d: %v", item.ProductID, err), http.StatusBadRequest)
			return
		}
	}

	order := &Order{
		CustomerID:  req.CustomerID,
		Items:       req.Items,
		TotalAmount: totalAmount,
	}

	createdOrder := store.Create(order)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdOrder)
}

func (h *Handler_Orders) Update_OrderStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	r_params := mux.Vars(r)
	orderID, err := strconv.Atoi(r_params["id"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var statusUpdate struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&statusUpdate); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// getting the controller's response
	err = h.controller.Update_OrderStatus(ctx, orderID, statusUpdate.Status)
	if err != nil {
		log.Printf("Error updating order status: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	// logging
	log.Printf("Order ID %d status updated to %s", orderID, statusUpdate.Status)
}
