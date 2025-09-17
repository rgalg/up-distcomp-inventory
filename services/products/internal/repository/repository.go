package products_repository

import (
	"context"
	"products-service/internal"
	dmodel "products-service/pkg"
	"sync"
)

// -------------------------------------------------------------------
// dtypes
// -------------------------------------------------------------------

// DataRepo_Products
// holds volatile data and mutex for concurrency
type DataRepo_Products struct {
	mu     sync.RWMutex
	items  map[int]*dmodel.Product
	nextID int
}

func New() *DataRepo_Products {
	datarepo := &DataRepo_Products{
		items:  make(map[int]*dmodel.Product),
		nextID: 1,
	}
	datarepo.addSampleProducts()
	return datarepo
}

// adding some sample products
func (dr *DataRepo_Products) addSampleProducts() {
	sampleProducts := []*dmodel.Product{
		{Name: "Laptop", Description: "High-performance laptop", Price: 999.99, Category: "Electronics"},
		{Name: "Mouse", Description: "Wireless optical mouse", Price: 29.99, Category: "Electronics"},
		{Name: "Keyboard", Description: "Mechanical keyboard", Price: 79.99, Category: "Electronics"},
		{Name: "Monitor", Description: "24-inch LCD monitor", Price: 199.99, Category: "Electronics"},
		{Name: "Desk Chair", Description: "Ergonomic office chair", Price: 149.99, Category: "Furniture"},
	}

	for _, product := range sampleProducts {
		product.ID = dr.nextID
		dr.items[dr.nextID] = product
		dr.nextID++
	}
}

// -------------------------------------------------------------------

// -------------------------------------------------------------------
// handling requests
// -------------------------------------------------------------------

// retrieving all items
func (dr *DataRepo_Products) Get_All(_ context.Context) ([]*dmodel.Product, error) {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	products := make([]*dmodel.Product, 0, len(dr.items))
	for _, product := range dr.items {
		products = append(products, product)
	}

	return products, nil
}

// retrieving item by ID
func (dr *DataRepo_Products) Get_ByProductID(_ context.Context, id int) (*dmodel.Product, error) {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	item, exists := dr.items[id]
	if !exists {
		return nil, internal.ErrItemNotFound
	}
	return item, nil
}

// creating a new product
func (dr *DataRepo_Products) Create_Product(_ context.Context, product *dmodel.Product) (*dmodel.Product, error) {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	product.ID = dr.nextID
	dr.items[dr.nextID] = product
	dr.nextID++

	return product, nil
}
