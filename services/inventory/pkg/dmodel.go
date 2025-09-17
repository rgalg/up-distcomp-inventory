package dmodel

type InventoryItem struct {
	ProductID int `json:"product_id"`
	Stock     int `json:"stock"`
	Reserved  int `json:"reserved"`
}
