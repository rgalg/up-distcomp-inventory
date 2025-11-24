package products_repository

import (
	"context"
	"database/sql"
	"products-service/internal"
	dmodel "products-service/pkg"
)

// -------------------------------------------------------------------
// dtypes
// -------------------------------------------------------------------

// DataRepo_Products
// data is in the DB now
type DataRepo_Products struct {
	db *sql.DB
}

func New(db *sql.DB) *DataRepo_Products {
	return &DataRepo_Products{
		db: db,
	}
}

// -------------------------------------------------------------------

// -------------------------------------------------------------------
// handling requests
// -------------------------------------------------------------------

// retrieving all items
func (dr *DataRepo_Products) Get_All(ctx context.Context) ([]*dmodel.Product, error) {
	query := `SELECT id, name, description, price, category FROM products`
	rows, err := dr.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*dmodel.Product
	for rows.Next() {
		var p dmodel.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Category); err != nil {
			return nil, err
		}
		products = append(products, &p)
	}

	return products, rows.Err()
}

// retrieving item by ID
func (dr *DataRepo_Products) Get_ByProductID(ctx context.Context, id int) (*dmodel.Product, error) {
	query := `SELECT id, name, description, price, category FROM products WHERE id = $1`
	var p dmodel.Product

	err := dr.db.QueryRowContext(ctx, query, id).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Category)
	if err == sql.ErrNoRows {
		return nil, internal.ErrItemNotFound
	}
	if err != nil {
		return nil, err
	}

	return &p, nil
}

// creating a new product
func (dr *DataRepo_Products) Create_Product(ctx context.Context, product *dmodel.Product) (*dmodel.Product, error) {
	query := `INSERT INTO products (name, description, price, category) VALUES ($1, $2, $3, $4) RETURNING id`

	err := dr.db.QueryRowContext(ctx, query, product.Name, product.Description, product.Price, product.Category).Scan(&product.ID)
	if err != nil {
		return nil, err
	}

	return product, nil
}
