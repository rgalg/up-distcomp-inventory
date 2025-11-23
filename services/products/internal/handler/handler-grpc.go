package products_handler_http

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	internal "products-service/internal"
	products_controller "products-service/internal/controller"
	products_dmodel "products-service/pkg"
	pb "products-service/proto/products"
)

type Handler_Products_GRPC struct {
	pb.UnimplementedProductServiceServer
	controller *products_controller.Controller_Products
}

func NewGRPC(controller *products_controller.Controller_Products) *Handler_Products_GRPC {
	return &Handler_Products_GRPC{
		controller: controller,
	}
}

func (h *Handler_Products_GRPC) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	product, err := h.controller.Get_ByProductID(ctx, int(req.Id))
	if err != nil {
		if err == internal.ErrItemNotFound {
			return nil, status.Errorf(codes.NotFound, "product not found")
		}
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	return &pb.GetProductResponse{
		Product: &pb.Product{
			Id:          int32(product.ID),
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Category:    product.Category,
		},
	}, nil
}

func (h *Handler_Products_GRPC) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	products, err := h.controller.Get_All(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	pbProducts := make([]*pb.Product, len(products))
	for i, product := range products {
		pbProducts[i] = &pb.Product{
			Id:          int32(product.ID),
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Category:    product.Category,
		}
	}

	return &pb.ListProductsResponse{
		Products: pbProducts,
	}, nil
}

func (h *Handler_Products_GRPC) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	product := &products_dmodel.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
	}

	createdProduct, err := h.controller.Create_Product(ctx, product)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	return &pb.CreateProductResponse{
		Product: &pb.Product{
			Id:          int32(createdProduct.ID),
			Name:        createdProduct.Name,
			Description: createdProduct.Description,
			Price:       createdProduct.Price,
			Category:    createdProduct.Category,
		},
	}, nil
}
