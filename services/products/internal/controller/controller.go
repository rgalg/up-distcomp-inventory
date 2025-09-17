package products_controller

import (
	"context"

	dmodel "products-service/pkg"
)

type if_repo_inventory interface {
	Get_All(_ context.Context) ([]*dmodel.Product, error)
	Get_ByProductID(_ context.Context, productID int) (*dmodel.Product, error)
	Create_Product(_ context.Context, product *dmodel.Product) (*dmodel.Product, error)
}

type Controller_Products struct {
	repo if_repo_inventory
}

func New(repo if_repo_inventory) *Controller_Products {
	return &Controller_Products{
		repo: repo,
	}
}

func (c *Controller_Products) Get_All(ctx context.Context) ([]*dmodel.Product, error) {
	res, err := c.repo.Get_All(ctx)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Controller_Products) Get_ByProductID(ctx context.Context, productID int) (*dmodel.Product, error) {
	res, err := c.repo.Get_ByProductID(ctx, productID)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Controller_Products) Create_Product(ctx context.Context, product *dmodel.Product) (*dmodel.Product, error) {
	res, err := c.repo.Create_Product(ctx, product)

	if err != nil {
		return nil, err
	}

	return res, nil
}
