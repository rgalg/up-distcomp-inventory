package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type InventoryItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
	Reserved  int `json:"reserved"`
}

// InventoryStore holds inventory items and a mutex for concurrency
type InventoryStore struct {
	mu    sync.RWMutex
	items map[int]*InventoryItem
}

// NewInventoryStore creates and returns a new InventoryStore with sample data
func NewInventoryStore() *InventoryStore {
	store := &InventoryStore{
		items: make(map[int]*InventoryItem),
	}
	store.addSampleInventory()
	return store
}

func (is *InventoryStore) addSampleInventory() {
	sampleItems := []*InventoryItem{
		{ProductID: 1, Quantity: 50, Reserved: 5},
		{ProductID: 2, Quantity: 100, Reserved: 0},
		{ProductID: 3, Quantity: 25, Reserved: 2},
		{ProductID: 4, Quantity: 30, Reserved: 0},
		{ProductID: 5, Quantity: 15, Reserved: 1},
	}

	for _, item := range sampleItems {
		is.items[item.ProductID] = item
	}
}

func (is *InventoryStore) GetAll() []*InventoryItem {
	is.mu.RLock()
	defer is.mu.RUnlock()

	items := make([]*InventoryItem, 0, len(is.items))
	for _, item := range is.items {
		items = append(items, item)
	}
	return items
}

func (is *InventoryStore) GetByProductID(productID int) (*InventoryItem, bool) {
	is.mu.RLock()
	defer is.mu.RUnlock()

	item, exists := is.items[productID]
	return item, exists
}

func (is *InventoryStore) UpdateQuantity(productID, quantity int) error {
	is.mu.Lock()
	defer is.mu.Unlock()

	item, exists := is.items[productID]
	if !exists {
		is.items[productID] = &InventoryItem{
			ProductID: productID,
			Quantity:  quantity,
			Reserved:  0,
		}
		return nil
	}

	item.Quantity = quantity
	return nil
}

func (is *InventoryStore) ReserveQuantity(productID, quantity int) error {
	is.mu.Lock()
	defer is.mu.Unlock()

	item, exists := is.items[productID]
	if !exists {
		return fmt.Errorf("product not found in inventory")
	}

	availableQuantity := item.Quantity - item.Reserved
	if availableQuantity < quantity {
		return fmt.Errorf("insufficient quantity available")
	}

	item.Reserved += quantity
	return nil
}

func (is *InventoryStore) ReleaseReservation(productID, quantity int) error {
	is.mu.Lock()
	defer is.mu.Unlock()

	item, exists := is.items[productID]
	if !exists {
		return fmt.Errorf("product not found in inventory")
	}

	if item.Reserved < quantity {
		return fmt.Errorf("cannot release more than reserved")
	}

	item.Reserved -= quantity
	return nil
}

func (is *InventoryStore) FulfillReservation(productID, quantity int) error {
	is.mu.Lock()
	defer is.mu.Unlock()

	item, exists := is.items[productID]
	if !exists {
		return fmt.Errorf("product not found in inventory")
	}

	if item.Reserved < quantity {
		return fmt.Errorf("cannot fulfill more than reserved")
	}

	item.Reserved -= quantity
	item.Quantity -= quantity
	return nil
}

var store *InventoryStore

func getAllInventory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	items := store.GetAll()
	json.NewEncoder(w).Encode(items)
}

func getInventoryByProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["productId"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	item, exists := store.GetByProductID(productID)
	if !exists {
		http.Error(w, "Product not found in inventory", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(item)
}

type UpdateQuantityRequest struct {
	Quantity int `json:"quantity"`
}

func updateInventory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["productId"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var req UpdateQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := store.UpdateQuantity(productID, req.Quantity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	item, _ := store.GetByProductID(productID)
	json.NewEncoder(w).Encode(item)
}

type ReserveRequest struct {
	Quantity int `json:"quantity"`
}

func reserveInventory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["productId"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var req ReserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := store.ReserveQuantity(productID, req.Quantity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Reservation successful")
}

func fulfillReservation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["productId"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var req ReserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := store.FulfillReservation(productID, req.Quantity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Fulfillment successful")
}

func corsHandler(next http.Handler) http.Handler {
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

func main() {
	// -------------------------------------------------------------------
	// data (volatile for now)
	// -------------------------------------------------------------------
	store = NewInventoryStore()
	// -------------------------------------------------------------------

	// -------------------------------------------------------------------
	// router (handler)
	// -------------------------------------------------------------------
	r := mux.NewRouter()
	r.Use(corsHandler)
	// -------------------------------------------------------------------

	// -------------------------------------------------------------------
	// service endpoints
	// -------------------------------------------------------------------
	// GET requests
	r.HandleFunc("/inventory", getAllInventory).Methods(http.MethodGet)
	r.HandleFunc("/inventory/{productId}", getInventoryByProduct).Methods(http.MethodGet)
	// PUT requests
	r.HandleFunc("/inventory/{productId}", updateInventory).Methods(http.MethodPut)
	// POST requests
	r.HandleFunc("/inventory/{productId}/reserve", reserveInventory).Methods(http.MethodPost)
	r.HandleFunc("/inventory/{productId}/fulfill", fulfillReservation).Methods(http.MethodPost)
	// -------------------------------------------------------------------
	// CORS preflight (OPTIONS) requests for all endpoints
	r.HandleFunc("/inventory", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodOptions)
	r.HandleFunc("/inventory/{productId}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodOptions)
	r.HandleFunc("/inventory/{productId}/reserve", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodOptions)
	r.HandleFunc("/inventory/{productId}/fulfill", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodOptions)
	// -------------------------------------------------------------------
	// health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Inventory service is healthy")
	}).Methods(http.MethodGet)
	// -------------------------------------------------------------------

	// -------------------------------------------------------------------
	// exposing the service
	// -------------------------------------------------------------------
	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}
	log.Printf("Inventory service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r)) // log and os.Exit(1)
	// -------------------------------------------------------------------
}
