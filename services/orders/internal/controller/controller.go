package orders_controller

import (
	"context"

	orders_dmodel "orders-service/pkg"
)

type if_repo_orders interface {
	Get_All(_ context.Context) ([]*orders_dmodel.Order, error)
	Get_ByOrderID(_ context.Context, id int) (*orders_dmodel.Order, error)
	Create_Order(_ context.Context, order *orders_dmodel.Order) (*orders_dmodel.Order, error)
	Update_OrderStatus(_ context.Context, orderID int, status string) error
}

type Controller_Orders struct {
	repo if_repo_orders
}

func New(repo if_repo_orders) *Controller_Orders {
	return &Controller_Orders{
		repo: repo,
	}
}

func (c *Controller_Orders) Get_All(ctx context.Context) ([]*orders_dmodel.Order, error) {
	res, err := c.repo.Get_All(ctx)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Controller_Orders) Get_ByOrderID(ctx context.Context, id int) (*orders_dmodel.Order, error) {
	res, err := c.repo.Get_ByOrderID(ctx, id)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Controller_Orders) Create_Order(ctx context.Context, order *orders_dmodel.Order) (*orders_dmodel.Order, error) {
	res, err := c.repo.Create_Order(ctx, order)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Controller_Orders) Update_OrderStatus(ctx context.Context, orderID int, status string) error {
	err := c.repo.Update_OrderStatus(ctx, orderID, status)

	if err != nil {
		return err
	}

	return nil
}
