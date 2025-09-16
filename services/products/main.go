package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
}

type ProductStore struct {
	mu       sync.RWMutex
	products map[int]*Product
	nextID   int
}

func NewProductStore() *ProductStore {
	store := &ProductStore{
		products: make(map[int]*Product),
		nextID:   1,
	}
	
	// Add some sample products
	store.addSampleProducts()
	return store
}

func (ps *ProductStore) addSampleProducts() {
	sampleProducts := []*Product{
		{Name: "Laptop", Description: "High-performance laptop", Price: 999.99, Category: "Electronics"},
		{Name: "Mouse", Description: "Wireless optical mouse", Price: 29.99, Category: "Electronics"},
		{Name: "Keyboard", Description: "Mechanical keyboard", Price: 79.99, Category: "Electronics"},
		{Name: "Monitor", Description: "24-inch LCD monitor", Price: 199.99, Category: "Electronics"},
		{Name: "Desk Chair", Description: "Ergonomic office chair", Price: 149.99, Category: "Furniture"},
	}
	
	for _, product := range sampleProducts {
		product.ID = ps.nextID
		ps.products[ps.nextID] = product
		ps.nextID++
	}
}

func (ps *ProductStore) GetAll() []*Product {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	
	products := make([]*Product, 0, len(ps.products))
	for _, product := range ps.products {
		products = append(products, product)
	}
	return products
}

func (ps *ProductStore) GetByID(id int) (*Product, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	
	product, exists := ps.products[id]
	return product, exists
}

func (ps *ProductStore) Create(product *Product) *Product {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	product.ID = ps.nextID
	ps.products[ps.nextID] = product
	ps.nextID++
	return product
}

var store *ProductStore

func getAllProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	products := store.GetAll()
	json.NewEncoder(w).Encode(products)
}

func getProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	
	product, exists := store.GetByID(id)
	if !exists {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}
	
	json.NewEncoder(w).Encode(product)
}

func createProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	var product Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	createdProduct := store.Create(&product)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdProduct)
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
	store = NewProductStore()
	
	r := mux.NewRouter()
	r.Use(corsHandler)
	
	r.HandleFunc("/products", getAllProducts).Methods("GET")
	r.HandleFunc("/products/{id}", getProduct).Methods("GET")
	r.HandleFunc("/products", createProduct).Methods("POST")
	
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Products service is healthy")
	}).Methods("GET")
	
	fmt.Println("Products service starting on port 8001...")
	log.Fatal(http.ListenAndServe(":8001", r))
}