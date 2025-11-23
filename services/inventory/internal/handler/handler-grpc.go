package inventory_handler_http

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	internal "inventory-service/internal"
	inventory_controller "inventory-service/internal/controller"
	pb "inventory-service/proto/inventory"
)

type Handler_Inventory_GRPC struct {
	pb.UnimplementedInventoryServiceServer
	controller *inventory_controller.Controller_Inventory
}

func NewGRPC(controller *inventory_controller.Controller_Inventory) *Handler_Inventory_GRPC {
	return &Handler_Inventory_GRPC{
		controller: controller,
	}
}

func (h *Handler_Inventory_GRPC) GetInventory(ctx context.Context, req *pb.GetInventoryRequest) (*pb.GetInventoryResponse, error) {
	item, err := h.controller.Get_ByProductID(ctx, int(req.ProductId))
	if err != nil {
		if err == internal.ErrItemNotFound {
			return nil, status.Errorf(codes.NotFound, "inventory not found")
		}
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	return &pb.GetInventoryResponse{
		Item: &pb.InventoryItem{
			ProductId: int32(item.ProductID),
			Stock:     int32(item.Stock),
			Reserved:  int32(item.Reserved),
		},
	}, nil
}

func (h *Handler_Inventory_GRPC) ListInventory(ctx context.Context, req *pb.ListInventoryRequest) (*pb.ListInventoryResponse, error) {
	items, err := h.controller.Get_All(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	pbItems := make([]*pb.InventoryItem, len(items))
	for i, item := range items {
		pbItems[i] = &pb.InventoryItem{
			ProductId: int32(item.ProductID),
			Stock:     int32(item.Stock),
			Reserved:  int32(item.Reserved),
		}
	}

	return &pb.ListInventoryResponse{
		Items: pbItems,
	}, nil
}

func (h *Handler_Inventory_GRPC) UpdateStock(ctx context.Context, req *pb.UpdateStockRequest) (*pb.UpdateStockResponse, error) {
	err := h.controller.Update_Stock(ctx, int(req.ProductId), int(req.Quantity))
	if err != nil {
		if err == internal.ErrItemNotFound {
			return nil, status.Errorf(codes.NotFound, "inventory not found")
		}
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	// Get the updated item
	updatedItem, err := h.controller.Get_ByProductID(ctx, int(req.ProductId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	return &pb.UpdateStockResponse{
		Item: &pb.InventoryItem{
			ProductId: int32(updatedItem.ProductID),
			Stock:     int32(updatedItem.Stock),
			Reserved:  int32(updatedItem.Reserved),
		},
	}, nil
}

func (h *Handler_Inventory_GRPC) ReserveStock(ctx context.Context, req *pb.ReserveStockRequest) (*pb.ReserveStockResponse, error) {
	err := h.controller.Reserve_Stock(ctx, int(req.ProductId), int(req.Stock))
	if err != nil {
		if err == internal.ErrItemNotFound {
			return nil, status.Errorf(codes.NotFound, "inventory not found")
		}
		if err == internal.ErrInsufficientStock {
			return nil, status.Errorf(codes.FailedPrecondition, "insufficient stock")
		}
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	// Get the updated item
	item, err := h.controller.Get_ByProductID(ctx, int(req.ProductId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	return &pb.ReserveStockResponse{
		Item: &pb.InventoryItem{
			ProductId: int32(item.ProductID),
			Stock:     int32(item.Stock),
			Reserved:  int32(item.Reserved),
		},
	}, nil
}

func (h *Handler_Inventory_GRPC) FulfillReservation(ctx context.Context, req *pb.FulfillReservationRequest) (*pb.FulfillReservationResponse, error) {
	err := h.controller.Fulfill_Reservation(ctx, int(req.ProductId), int(req.Stock))
	if err != nil {
		if err == internal.ErrItemNotFound {
			return nil, status.Errorf(codes.NotFound, "inventory not found")
		}
		if err == internal.ErrInsufficientReserved {
			return nil, status.Errorf(codes.FailedPrecondition, "insufficient reserved stock")
		}
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	// Get the updated item
	item, err := h.controller.Get_ByProductID(ctx, int(req.ProductId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	return &pb.FulfillReservationResponse{
		Item: &pb.InventoryItem{
			ProductId: int32(item.ProductID),
			Stock:     int32(item.Stock),
			Reserved:  int32(item.Reserved),
		},
	}, nil
}

func (h *Handler_Inventory_GRPC) ReleaseReservation(ctx context.Context, req *pb.ReleaseReservationRequest) (*pb.ReleaseReservationResponse, error) {
	err := h.controller.Release_Reservation(ctx, int(req.ProductId), int(req.Stock))
	if err != nil {
		if err == internal.ErrItemNotFound {
			return nil, status.Errorf(codes.NotFound, "inventory not found")
		}
		if err == internal.ErrInsufficientReserved {
			return nil, status.Errorf(codes.FailedPrecondition, "insufficient reserved stock")
		}
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	// Get the updated item
	item, err := h.controller.Get_ByProductID(ctx, int(req.ProductId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	return &pb.ReleaseReservationResponse{
		Item: &pb.InventoryItem{
			ProductId: int32(item.ProductID),
			Stock:     int32(item.Stock),
			Reserved:  int32(item.Reserved),
		},
	}, nil
}
