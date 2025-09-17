package dmodel

import "time"

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
