package inventory_controller

import (
	"context"
	dmodel "inventory-service/pkg"
)

type if_repo_inventory interface {
	Get_All(_ context.Context) ([]*dmodel.InventoryItem, error)
	Get_ByProductID(_ context.Context, productID int) (*dmodel.InventoryItem, error)
	Update_Stock(_ context.Context, productID, stock int) error
	Reserve_Stock(_ context.Context, productID, amount_reserved int) error
	Release_Reservation(_ context.Context, productID, amount_released int) error
	Fulfill_Reservation(_ context.Context, productID, amount_fulfilled int) error
}

type Controller_Inventory struct {
	repo if_repo_inventory
}

func New(repo if_repo_inventory) *Controller_Inventory {
	return &Controller_Inventory{
		repo: repo,
	}
}

func (c *Controller_Inventory) Get_All(ctx context.Context) ([]*dmodel.InventoryItem, error) {
	res, err := c.repo.Get_All(ctx)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Controller_Inventory) Get_ByProductID(ctx context.Context, productID int) (*dmodel.InventoryItem, error) {
	res, err := c.repo.Get_ByProductID(ctx, productID)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Controller_Inventory) Update_Stock(ctx context.Context, productID, stock int) error {
	err := c.repo.Update_Stock(ctx, productID, stock)

	if err != nil {
		return err
	}

	return nil
}

func (c *Controller_Inventory) Reserve_Stock(ctx context.Context, productID, amount_reserved int) error {
	err := c.repo.Reserve_Stock(ctx, productID, amount_reserved)

	if err != nil {
		return err
	}

	return nil
}

func (c *Controller_Inventory) Release_Reservation(ctx context.Context, productID, amount_released int) error {
	err := c.repo.Release_Reservation(ctx, productID, amount_released)

	if err != nil {
		return err
	}

	return nil
}

func (c *Controller_Inventory) Fulfill_Reservation(ctx context.Context, productID, amount_fulfilled int) error {
	err := c.repo.Fulfill_Reservation(ctx, productID, amount_fulfilled)

	if err != nil {
		return err
	}

	return nil
}
