package inventory_repository

import (
	"context"
	"fmt"
	"sync"

	internal "inventory-service/internal"
	dmodel "inventory-service/pkg"
)

// -------------------------------------------------------------------
// dtypes
// -------------------------------------------------------------------

// DataRepo_Inventory
// holds volatile data and a mutex for concurrency
type DataRepo_Inventory struct {
	mu    sync.RWMutex
	items map[int]*dmodel.InventoryItem
}

// create a new object with mock data
func New() *DataRepo_Inventory {
	datarepo := &DataRepo_Inventory{
		items: make(map[int]*dmodel.InventoryItem),
	}
	datarepo.addSampleInventory()
	return datarepo
}

// adding some sample inventory items
func (dr *DataRepo_Inventory) addSampleInventory() {
	sampleItems := []*dmodel.InventoryItem{
		{ProductID: 1, Stock: 50, Reserved: 0},
		{ProductID: 2, Stock: 100, Reserved: 0},
		{ProductID: 3, Stock: 25, Reserved: 0},
		{ProductID: 4, Stock: 30, Reserved: 0},
		{ProductID: 5, Stock: 15, Reserved: 0},
	}

	for _, item := range sampleItems {
		dr.items[item.ProductID] = item
	}
}

// -------------------------------------------------------------------

// -------------------------------------------------------------------
// handling requests
// -------------------------------------------------------------------

// retrieving all items
func (dr *DataRepo_Inventory) Get_All(_ context.Context) ([]*dmodel.InventoryItem, error) {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	items := make([]*dmodel.InventoryItem, 0, len(dr.items))
	for _, item := range dr.items {
		items = append(items, item)
	}

	return items, nil
}

// retrieving item by product ID
func (dr *DataRepo_Inventory) Get_ByProductID(_ context.Context, productID int) (*dmodel.InventoryItem, error) {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	item, exists := dr.items[productID]
	if !exists {
		return nil, internal.ErrItemNotFound
	}

	return item, nil
}

// -------------------------------------------------------------------

// update the stock property of an inventory item
func (dr *DataRepo_Inventory) Update_Stock(_ context.Context, productID, stock int) error {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	item, exists := dr.items[productID]

	if !exists {
		return internal.ErrItemNotFound
	}

	item.Stock = stock
	return nil
}

// increase the reserved property of an item
func (dr *DataRepo_Inventory) Reserve_Stock(_ context.Context, productID, amount_reserved int) error {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	item, exists := dr.items[productID]

	if !exists {
		return internal.ErrItemNotFound
	}
	if (item.Stock - item.Reserved) < amount_reserved {
		return internal.ErrInsufficientStock
	}

	item.Reserved += amount_reserved
	return nil
}

// decrease the reserved property of an item
func (dr *DataRepo_Inventory) Release_Reservation(_ context.Context, productID, amount_released int) error {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	item, exists := dr.items[productID]
	if !exists {
		return internal.ErrItemNotFound
	}

	if item.Reserved < amount_released {
		return internal.ErrInsufficientReserved
	}

	item.Reserved -= amount_released
	return nil
}

// decrease both the reserved and quantity properties of an item
// used to fulfill an order
func (dr *DataRepo_Inventory) Fulfill_Reservation(_ context.Context, productID, amount_fulfilled int) error {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	item, exists := dr.items[productID]
	if !exists {
		return fmt.Errorf("product not found in inventory")
	}

	if item.Reserved < amount_fulfilled {
		return fmt.Errorf("cannot fulfill more items than are reserved")
	}

	item.Reserved -= amount_fulfilled
	item.Stock -= amount_fulfilled
	return nil
}

// -------------------------------------------------------------------
