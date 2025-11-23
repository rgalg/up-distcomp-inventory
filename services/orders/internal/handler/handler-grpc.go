package orders_handler_http

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	internal "orders-service/internal"
	orders_controller "orders-service/internal/controller"
	orders_dmodel "orders-service/pkg"
	pb "orders-service/proto/orders"
	
	products_pb "orders-service/proto/products"
	inventory_pb "orders-service/proto/inventory"
)

type Handler_Orders_GRPC struct {
	pb.UnimplementedOrderServiceServer
	controller        *orders_controller.Controller_Orders
	productsClient    products_pb.ProductServiceClient
	inventoryClient   inventory_pb.InventoryServiceClient
}

func NewGRPC(controller *orders_controller.Controller_Orders, productsAddr, inventoryAddr string) (*Handler_Orders_GRPC, error) {
	// Connect to Products service
	productsConn, err := grpc.NewClient(productsAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	productsClient := products_pb.NewProductServiceClient(productsConn)

	// Connect to Inventory service
	inventoryConn, err := grpc.NewClient(inventoryAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	inventoryClient := inventory_pb.NewInventoryServiceClient(inventoryConn)

	return &Handler_Orders_GRPC{
		controller:      controller,
		productsClient:  productsClient,
		inventoryClient: inventoryClient,
	}, nil
}

func (h *Handler_Orders_GRPC) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	order, err := h.controller.Get_ByOrderID(ctx, int(req.Id))
	if err != nil {
		if err == internal.ErrItemNotFound {
			return nil, status.Errorf(codes.NotFound, "order not found")
		}
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	pbItems := make([]*pb.OrderItem, len(order.Items))
	for i, item := range order.Items {
		pbItems[i] = &pb.OrderItem{
			ProductId: int32(item.ProductID),
			Quantity:  int32(item.Quantity),
		}
	}

	return &pb.GetOrderResponse{
		Order: &pb.Order{
			Id:          int32(order.ID),
			CustomerId:  int32(order.CustomerID),
			Items:       pbItems,
			Status:      order.Status,
			TotalAmount: order.TotalAmount,
			CreatedAt:   order.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (h *Handler_Orders_GRPC) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	orders, err := h.controller.Get_All(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	pbOrders := make([]*pb.Order, len(orders))
	for i, order := range orders {
		pbItems := make([]*pb.OrderItem, len(order.Items))
		for j, item := range order.Items {
			pbItems[j] = &pb.OrderItem{
				ProductId: int32(item.ProductID),
				Quantity:  int32(item.Quantity),
			}
		}

		pbOrders[i] = &pb.Order{
			Id:          int32(order.ID),
			CustomerId:  int32(order.CustomerID),
			Items:       pbItems,
			Status:      order.Status,
			TotalAmount: order.TotalAmount,
			CreatedAt:   order.CreatedAt.Format(time.RFC3339),
		}
	}

	return &pb.ListOrdersResponse{
		Orders: pbOrders,
	}, nil
}

func (h *Handler_Orders_GRPC) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	// Validate items
	if len(req.Items) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "order must contain at least one item")
	}

	// Calculate total amount and validate products
	var totalAmount float64
	items := make([]orders_dmodel.OrderItem, len(req.Items))
	for i, item := range req.Items {
		// Get product details from Products service via gRPC
		productResp, err := h.productsClient.GetProduct(ctx, &products_pb.GetProductRequest{
			Id: item.ProductId,
		})
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return nil, status.Errorf(codes.InvalidArgument, "product %d not found", item.ProductId)
			}
			return nil, status.Errorf(codes.Internal, "failed to get product: %v", err)
		}

		totalAmount += productResp.Product.Price * float64(item.Quantity)

		// Reserve inventory via gRPC
		_, err = h.inventoryClient.ReserveStock(ctx, &inventory_pb.ReserveStockRequest{
			ProductId: item.ProductId,
			Stock:     item.Quantity,
		})
		if err != nil {
			log.Printf("Failed to reserve inventory for product %d: %v", item.ProductId, err)
			return nil, status.Errorf(codes.FailedPrecondition, "failed to reserve inventory for product %d: %v", item.ProductId, err)
		}

		items[i] = orders_dmodel.OrderItem{
			ProductID: int(item.ProductId),
			Quantity:  int(item.Quantity),
		}
	}

	order := &orders_dmodel.Order{
		CustomerID:  int(req.CustomerId),
		Items:       items,
		TotalAmount: totalAmount,
	}

	createdOrder, err := h.controller.Create_Order(ctx, order)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	pbItems := make([]*pb.OrderItem, len(createdOrder.Items))
	for i, item := range createdOrder.Items {
		pbItems[i] = &pb.OrderItem{
			ProductId: int32(item.ProductID),
			Quantity:  int32(item.Quantity),
		}
	}

	return &pb.CreateOrderResponse{
		Order: &pb.Order{
			Id:          int32(createdOrder.ID),
			CustomerId:  int32(createdOrder.CustomerID),
			Items:       pbItems,
			Status:      createdOrder.Status,
			TotalAmount: createdOrder.TotalAmount,
			CreatedAt:   createdOrder.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (h *Handler_Orders_GRPC) FulfillOrder(ctx context.Context, req *pb.FulfillOrderRequest) (*pb.FulfillOrderResponse, error) {
	order, err := h.controller.Get_ByOrderID(ctx, int(req.Id))
	if err != nil {
		if err == internal.ErrItemNotFound {
			return nil, status.Errorf(codes.NotFound, "order not found")
		}
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	if order.Status != "pending" {
		return nil, status.Errorf(codes.FailedPrecondition, "order is not in pending status")
	}

	// Fulfill inventory for each item via gRPC
	for _, item := range order.Items {
		_, err := h.inventoryClient.FulfillReservation(ctx, &inventory_pb.FulfillReservationRequest{
			ProductId: int32(item.ProductID),
			Stock:     int32(item.Quantity),
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to fulfill inventory: %v", err)
		}
	}

	// Update order status
	if err := h.controller.Update_OrderStatus(ctx, int(req.Id), "fulfilled"); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update order status")
	}

	updatedOrder, err := h.controller.Get_ByOrderID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	pbItems := make([]*pb.OrderItem, len(updatedOrder.Items))
	for i, item := range updatedOrder.Items {
		pbItems[i] = &pb.OrderItem{
			ProductId: int32(item.ProductID),
			Quantity:  int32(item.Quantity),
		}
	}

	return &pb.FulfillOrderResponse{
		Order: &pb.Order{
			Id:          int32(updatedOrder.ID),
			CustomerId:  int32(updatedOrder.CustomerID),
			Items:       pbItems,
			Status:      updatedOrder.Status,
			TotalAmount: updatedOrder.TotalAmount,
			CreatedAt:   updatedOrder.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}
