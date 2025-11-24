package orders_repository

import (
	"context"
	"database/sql"
	internal "orders-service/internal"
	dmodel "orders-service/pkg"
	"time"
)

// -------------------------------------------------------------------
// dtypes
// -------------------------------------------------------------------

// DataRepo_Orders
// data in a separate DB now
type DataRepo_Orders struct {
	db *sql.DB
}

func New(db *sql.DB) *DataRepo_Orders {
	return &DataRepo_Orders{
		db: db,
	}
}

// -------------------------------------------------------------------

// -------------------------------------------------------------------
// handling requests
// -------------------------------------------------------------------

func (dr *DataRepo_Orders) Get_All(ctx context.Context) ([]*dmodel.Order, error) {
	query := `SELECT id, customer_id, status, total_amount, created_at FROM orders ORDER BY created_at DESC`
	rows, err := dr.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*dmodel.Order
	for rows.Next() {
		var o dmodel.Order
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.Status, &o.TotalAmount, &o.CreatedAt); err != nil {
			return nil, err
		}

		// Load order items
		items, err := dr.getOrderItems(ctx, o.ID)
		if err != nil {
			return nil, err
		}
		o.Items = items

		orders = append(orders, &o)
	}

	return orders, rows.Err()
}

func (dr *DataRepo_Orders) Get_ByOrderID(ctx context.Context, id int) (*dmodel.Order, error) {
	query := `SELECT id, customer_id, status, total_amount, created_at FROM orders WHERE id = $1`
	var o dmodel.Order

	err := dr.db.QueryRowContext(ctx, query, id).Scan(&o.ID, &o.CustomerID, &o.Status, &o.TotalAmount, &o.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, internal.ErrItemNotFound
	}
	if err != nil {
		return nil, err
	}

	// Load order items
	items, err := dr.getOrderItems(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return &o, nil
}

// -------------------------------------------------------------------

func (dr *DataRepo_Orders) getOrderItems(ctx context.Context, orderID int) ([]dmodel.OrderItem, error) {
	query := `SELECT product_id, quantity, price_at_order FROM order_items WHERE order_id = $1`
	rows, err := dr.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []dmodel.OrderItem
	for rows.Next() {
		var item dmodel.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.Price); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// -------------------------------------------------------------------

func (dr *DataRepo_Orders) Create_Order(ctx context.Context, order *dmodel.Order) (*dmodel.Order, error) {
	tx, err := dr.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	order.CreatedAt = time.Now()
	order.Status = "pending"

	query := `INSERT INTO orders (customer_id, status, total_amount, created_at) VALUES ($1, $2, $3, $4) RETURNING id`
	err = tx.QueryRowContext(ctx, query, order.CustomerID, order.Status, order.TotalAmount, order.CreatedAt).Scan(&order.ID)
	if err != nil {
		return nil, err
	}

	// Insert order items
	itemQuery := `INSERT INTO order_items (order_id, product_id, quantity, price_at_order) VALUES ($1, $2, $3, $4)`
	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx, itemQuery, order.ID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return order, nil
}

// -------------------------------------------------------------------

func (dr *DataRepo_Orders) Update_OrderStatus(ctx context.Context, id int, status string) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`
	result, err := dr.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return internal.ErrItemNotFound
	}

	return nil
}

// -------------------------------------------------------------------
