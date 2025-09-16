package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type OrderItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type Order struct {
	ID          int         `json:"id"`
	CustomerID  int         `json:"customer_id"`
	Items       []OrderItem `json:"items"`
	Status      string      `json:"status"`
	TotalAmount float64     `json:"total_amount"`
	CreatedAt   time.Time   `json:"created_at"`
}

type OrderStore struct {
	mu     sync.RWMutex
	orders map[int]*Order
	nextID int
}

func NewOrderStore() *OrderStore {
	return &OrderStore{
		orders: make(map[int]*Order),
		nextID: 1,
	}
}

func (os *OrderStore) GetAll() []*Order {
	os.mu.RLock()
	defer os.mu.RUnlock()
	
	orders := make([]*Order, 0, len(os.orders))
	for _, order := range os.orders {
		orders = append(orders, order)
	}
	return orders
}

func (os *OrderStore) GetByID(id int) (*Order, bool) {
	os.mu.RLock()
	defer os.mu.RUnlock()
	
	order, exists := os.orders[id]
	return order, exists
}

func (os *OrderStore) Create(order *Order) *Order {
	os.mu.Lock()
	defer os.mu.Unlock()
	
	order.ID = os.nextID
	order.CreatedAt = time.Now()
	order.Status = "pending"
	os.orders[os.nextID] = order
	os.nextID++
	return order
}

func (os *OrderStore) UpdateStatus(id int, status string) error {
	os.mu.Lock()
	defer os.mu.Unlock()
	
	order, exists := os.orders[id]
	if !exists {
		return fmt.Errorf("order not found")
	}
	
	order.Status = status
	return nil
}

var store *OrderStore

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func getProduct(productID int) (*Product, error) {
	resp, err := http.Get(fmt.Sprintf("http://products:8001/products/%d", productID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product not found")
	}
	
	var product Product
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, err
	}
	
	return &product, nil
}

func reserveInventory(productID, quantity int) error {
	reqBody, _ := json.Marshal(map[string]int{"quantity": quantity})
	
	resp, err := http.Post(
		fmt.Sprintf("http://inventory:8002/inventory/%d/reserve", productID),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to reserve inventory")
	}
	
	return nil
}

func fulfillInventory(productID, quantity int) error {
	reqBody, _ := json.Marshal(map[string]int{"quantity": quantity})
	
	resp, err := http.Post(
		fmt.Sprintf("http://inventory:8002/inventory/%d/fulfill", productID),
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

func getAllOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	orders := store.GetAll()
	json.NewEncoder(w).Encode(orders)
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}
	
	order, exists := store.GetByID(id)
	if !exists {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	
	json.NewEncoder(w).Encode(order)
}

type CreateOrderRequest struct {
	CustomerID int         `json:"customer_id"`
	Items      []OrderItem `json:"items"`
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if len(req.Items) == 0 {
		http.Error(w, "Order must contain at least one item", http.StatusBadRequest)
		return
	}
	
	// Calculate total amount and validate products
	var totalAmount float64
	for _, item := range req.Items {
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

func fulfillOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}
	
	order, exists := store.GetByID(id)
	if !exists {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	
	if order.Status != "pending" {
		http.Error(w, "Order is not in pending status", http.StatusBadRequest)
		return
	}
	
	// Fulfill inventory for each item
	for _, item := range order.Items {
		if err := fulfillInventory(item.ProductID, item.Quantity); err != nil {
			http.Error(w, fmt.Sprintf("Failed to fulfill inventory: %v", err), http.StatusInternalServerError)
			return
		}
	}
	
	// Update order status
	if err := store.UpdateStatus(id, "fulfilled"); err != nil {
		http.Error(w, "Failed to update order status", http.StatusInternalServerError)
		return
	}
	
	updatedOrder, _ := store.GetByID(id)
	json.NewEncoder(w).Encode(updatedOrder)
}

func corsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func main() {
	store = NewOrderStore()
	
	r := mux.NewRouter()
	r.Use(corsHandler)
	
	r.HandleFunc("/orders", getAllOrders).Methods("GET")
	r.HandleFunc("/orders/{id}", getOrder).Methods("GET")
	r.HandleFunc("/orders", createOrder).Methods("POST")
	r.HandleFunc("/orders/{id}/fulfill", fulfillOrder).Methods("POST")
	
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Orders service is healthy")
	}).Methods("GET")
	
	fmt.Println("Orders service starting on port 8003...")
	log.Fatal(http.ListenAndServe(":8003", r))
}