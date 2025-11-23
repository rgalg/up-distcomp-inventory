package orders_handler_http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"

	internal "orders-service/internal"
	orders_controller "orders-service/internal/controller"
	orders_dmodel "orders-service/pkg"
	products_dmodel "orders-service/pkg/products"
)

func (h *Handler_Orders) getProduct(productID int) (*products_dmodel.Product, error) {
	// Use Kubernetes service discovery (environment variable or default)
	host := os.Getenv("PRODUCTS_HOST")
	if host == "" {
		host = "products-service:8001"
	}

	resp, err := http.Get(fmt.Sprintf("http://%s/products/%d", host, productID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product not found")
	}

	var product products_dmodel.Product
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, err
	}

	return &product, nil
}

func (h *Handler_Orders) reserveInventory(productID, quantity int) error {
	// Use Kubernetes service discovery (environment variable or default)
	host := os.Getenv("INVENTORY_HOST")
	if host == "" {
		host = "inventory-service:8002"
	}

	reqBody, _ := json.Marshal(map[string]int{"stock": quantity})

	resp, err := http.Post(
		fmt.Sprintf("http://%s/inventory/%d/reserve", host, productID),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Try to read backend error message for better debugging
		var backendMsg string
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		backendMsg = buf.String()
		if backendMsg == "" {
			backendMsg = resp.Status
		}
		return fmt.Errorf("failed to reserve inventory: %s", backendMsg)
	}

	return nil
}

func (h *Handler_Orders) fulfillInventory(productID, quantity int) error {
	// Use Kubernetes service discovery (environment variable or default)
	host := os.Getenv("INVENTORY_HOST")
	if host == "" {
		host = "inventory-service:8002"
	}

	reqBody, _ := json.Marshal(map[string]int{"stock": quantity})

	resp, err := http.Post(
		fmt.Sprintf("http://%s/inventory/%d/fulfill", host, productID),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fulfill inventory")
	}

	return nil
}

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

func New(controller *orders_controller.Controller_Orders, _ interface{}) *Handler_Orders {
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
	orderID, err := strconv.Atoi(r_params["orderId"])
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
		CustomerID int                       `json:"customer_id"`
		Items      []orders_dmodel.OrderItem `json:"items"`
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
		product, err := h.getProduct(item.ProductID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Product %d not found", item.ProductID), http.StatusBadRequest)
			return
		}

		totalAmount += product.Price * float64(item.Quantity)

		// Try to reserve inventory
		if err := h.reserveInventory(item.ProductID, item.Quantity); err != nil {
			http.Error(w, fmt.Sprintf("Failed to reserve inventory for product %d: %v", item.ProductID, err), http.StatusBadRequest)
			return
		}
	}

	order := &orders_dmodel.Order{
		CustomerID:  template_req.CustomerID,
		Items:       template_req.Items,
		TotalAmount: totalAmount,
	}

	createdOrder, err := h.controller.Create_Order(ctx, order)
	if err != nil {
		log.Printf("Error creating order: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdOrder)
}

func (h *Handler_Orders) Update_OrderStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	r_params := mux.Vars(r)
	orderID, err := strconv.Atoi(r_params["orderId"])
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

func (h *Handler_Orders) Fulfill_Order(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	r_params := mux.Vars(r)
	id, err := strconv.Atoi(r_params["orderId"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := h.controller.Get_ByOrderID(ctx, id)
	if err != nil {
		if err == internal.ErrItemNotFound {
			http.Error(w, "Order not found", http.StatusNotFound)
		} else {
			log.Printf("Error getting order by ID: Repository error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if order.Status != "pending" {
		http.Error(w, "Order is not in pending status", http.StatusBadRequest)
		return
	}

	// Fulfill inventory for each item
	for _, item := range order.Items {
		if err := h.fulfillInventory(item.ProductID, item.Quantity); err != nil {
			http.Error(w, fmt.Sprintf("Failed to fulfill inventory: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Update order status
	if err := h.controller.Update_OrderStatus(ctx, id, "fulfilled"); err != nil {
		http.Error(w, "Failed to update order status", http.StatusInternalServerError)
		return
	}

	updatedOrder, err := h.controller.Get_ByOrderID(ctx, id)
	if err != nil {
		log.Printf("Error getting updated order by ID: Repository error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updatedOrder)
}
