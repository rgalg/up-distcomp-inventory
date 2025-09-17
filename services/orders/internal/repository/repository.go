package orders_repository

import (
	"context"
	"sync"
	"time"

	internal "orders-service/internal"
	dmodel "orders-service/pkg"
)

// -------------------------------------------------------------------
// dtypes
// -------------------------------------------------------------------

// DataRepo_Orders
// holds volatile data and a mutex for concurrency
type DataRepo_Orders struct {
	mu     sync.RWMutex
	orders map[int]*dmodel.Order
	nextID int
}

func New() *DataRepo_Orders {
	return &DataRepo_Orders{
		orders: make(map[int]*dmodel.Order),
		nextID: 1,
	}
}

// -------------------------------------------------------------------

// -------------------------------------------------------------------
// handling requests
// -------------------------------------------------------------------

func (dr *DataRepo_Orders) Get_All(_ context.Context) ([]*dmodel.Order, error) {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	orders := make([]*dmodel.Order, 0, len(dr.orders))
	for _, order := range dr.orders {
		orders = append(orders, order)
	}

	return orders, nil
}

func (dr *DataRepo_Orders) Get_ByOrderID(_ context.Context, id int) (*dmodel.Order, error) {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	order, exists := dr.orders[id]
	if !exists {
		return nil, internal.ErrItemNotFound
	}

	return order, nil
}

// -------------------------------------------------------------------

func (dr *DataRepo_Orders) Create_Order(_ context.Context, order *dmodel.Order) (*dmodel.Order, error) {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	order.ID = dr.nextID
	order.CreatedAt = time.Now()
	order.Status = "pending"
	dr.orders[dr.nextID] = order
	dr.nextID++

	return order, nil
}

func (dr *DataRepo_Orders) Update_OrderStatus(_ context.Context, id int, status string) error {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	order, exists := dr.orders[id]
	if !exists {
		return internal.ErrItemNotFound
	}

	order.Status = status
	return nil
}

// -------------------------------------------------------------------
