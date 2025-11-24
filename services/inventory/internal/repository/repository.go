package inventory_repository

import (
	"context"
	"database/sql"
	"fmt"
	internal "inventory-service/internal"
	dmodel "inventory-service/pkg"
)

// -------------------------------------------------------------------
// dtypes
// -------------------------------------------------------------------

// DataRepo_Inventory
// holds volatile data and a mutex for concurrency
type DataRepo_Inventory struct {
	db *sql.DB
}

// create a new object with mock data
func New(db *sql.DB) *DataRepo_Inventory {
	return &DataRepo_Inventory{
		db: db,
	}
}

// -------------------------------------------------------------------

// -------------------------------------------------------------------
// handling requests
// -------------------------------------------------------------------

// retrieving all items
func (dr *DataRepo_Inventory) Get_All(ctx context.Context) ([]*dmodel.InventoryItem, error) {
	query := `SELECT product_id, stock, reserved FROM inventory`
	rows, err := dr.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*dmodel.InventoryItem
	for rows.Next() {
		var item dmodel.InventoryItem
		if err := rows.Scan(&item.ProductID, &item.Stock, &item.Reserved); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, rows.Err()
}

// retrieving item by product ID
func (dr *DataRepo_Inventory) Get_ByProductID(ctx context.Context, productID int) (*dmodel.InventoryItem, error) {
	query := `SELECT product_id, stock, reserved FROM inventory WHERE product_id = $1`
	var item dmodel.InventoryItem

	err := dr.db.QueryRowContext(ctx, query, productID).Scan(&item.ProductID, &item.Stock, &item.Reserved)
	if err == sql.ErrNoRows {
		return nil, internal.ErrItemNotFound
	}
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// -------------------------------------------------------------------

// update the stock property of an inventory item
func (dr *DataRepo_Inventory) Update_Stock(ctx context.Context, productID, stock int) error {
	query := `UPDATE inventory SET stock = $1, updated_at = CURRENT_TIMESTAMP WHERE product_id = $2`
	result, err := dr.db.ExecContext(ctx, query, stock, productID)
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

// increase the reserved property of an item
func (dr *DataRepo_Inventory) Reserve_Stock(ctx context.Context, productID, amount_reserved int) error {
	tx, err := dr.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var stock, reserved int
	query := `SELECT stock, reserved FROM inventory WHERE product_id = $1 FOR UPDATE`
	err = tx.QueryRowContext(ctx, query, productID).Scan(&stock, &reserved)
	if err == sql.ErrNoRows {
		return internal.ErrItemNotFound
	}
	if err != nil {
		return err
	}

	if (stock - reserved) < amount_reserved {
		return internal.ErrInsufficientStock
	}

	updateQuery := `UPDATE inventory SET reserved = reserved + $1, updated_at = CURRENT_TIMESTAMP WHERE product_id = $2`
	_, err = tx.ExecContext(ctx, updateQuery, amount_reserved, productID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// decrease the reserved property of an item
func (dr *DataRepo_Inventory) Release_Reservation(ctx context.Context, productID, amount_released int) error {
	tx, err := dr.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var reserved int
	query := `SELECT reserved FROM inventory WHERE product_id = $1 FOR UPDATE`
	err = tx.QueryRowContext(ctx, query, productID).Scan(&reserved)
	if err == sql.ErrNoRows {
		return internal.ErrItemNotFound
	}
	if err != nil {
		return err
	}

	if reserved < amount_released {
		return internal.ErrInsufficientReserved
	}

	updateQuery := `UPDATE inventory SET reserved = reserved - $1, updated_at = CURRENT_TIMESTAMP WHERE product_id = $2`
	_, err = tx.ExecContext(ctx, updateQuery, amount_released, productID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// decrease both the reserved and quantity properties of an item
// used to fulfill an order
func (dr *DataRepo_Inventory) Fulfill_Reservation(ctx context.Context, productID, amount_fulfilled int) error {
	tx, err := dr.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var reserved int
	query := `SELECT reserved FROM inventory WHERE product_id = $1 FOR UPDATE`
	err = tx.QueryRowContext(ctx, query, productID).Scan(&reserved)
	if err == sql.ErrNoRows {
		return fmt.Errorf("product not found in inventory")
	}
	if err != nil {
		return err
	}

	if reserved < amount_fulfilled {
		return fmt.Errorf("cannot fulfill more items than are reserved")
	}

	updateQuery := `UPDATE inventory SET reserved = reserved - $1, stock = stock - $1, updated_at = CURRENT_TIMESTAMP WHERE product_id = $2`
	_, err = tx.ExecContext(ctx, updateQuery, amount_fulfilled, productID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// -------------------------------------------------------------------
