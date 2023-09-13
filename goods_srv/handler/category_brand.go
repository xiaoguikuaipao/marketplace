package handler

import (
	"context"

	"grpc/goods_srv/global"
	"grpc/goods_srv/models"
	"grpc/goods_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *GoodsServer) CategoryBrandList(ctx context.Context, req *proto.CategoryBrandFilterRequest) (*proto.CategoryBrandListResponse, error) {
	var categoryBrands []models.GoodsCategoryBrand
	categoryBrandsListResponse := proto.CategoryBrandListResponse{}

	var total int64
	global.DB.Model(&models.GoodsCategoryBrand{}).Count(&total)
	categoryBrandsListResponse.Total = int32(total)

	//注意有外键需要预加载
	global.DB.Preload("Category").Preload("Brands").Scopes(paginate(int(req.Pages), int(req.PagePerNums))).Find(&categoryBrands)

	var categoryResponse []*proto.CategoryBrandResponse
	for _, categoryBrand := range categoryBrands {
		categoryResponse = append(categoryResponse, &proto.CategoryBrandResponse{
			Brand: &proto.BrandInfoResponse{
				Id:   categoryBrand.BrandsID,
				Name: categoryBrand.Category.Name,
				Logo: categoryBrand.Brands.Logo,
			},
			Category: &proto.CategoryInfoResponse{
				Id:             categoryBrand.Category.ID,
				Name:           categoryBrand.Category.Name,
				ParentCategory: categoryBrand.Category.ParentCategoryID,
				Level:          categoryBrand.Category.Level,
				IsTab:          categoryBrand.Category.IsTab,
			},
		})
	}
	categoryBrandsListResponse.Data = categoryResponse
	return &categoryBrandsListResponse, nil
}

// GetCategoryBrandList 根据分类列出该分类的所有品牌
func (s *GoodsServer) GetCategoryBrandList(ctx context.Context, req *proto.CategoryInfoRequest) (*proto.BrandListResponse, error) {
	brandListResponse := proto.BrandListResponse{}

	var category models.Category
	if result := global.DB.Find(&category, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "该分类不存在")
	}

	var categoryBrands []models.GoodsCategoryBrand
	if result := global.DB.Preload("Brands").Where(&models.GoodsCategoryBrand{CategoryID: category.ID}).Find(&categoryBrands); result.RowsAffected > 0 {
		brandListResponse.Total = int32(result.RowsAffected)
	}

	var brandInfoResponses []*proto.BrandInfoResponse
	for _, categoryBrand := range categoryBrands {
		brandInfoResponses = append(brandInfoResponses, &proto.BrandInfoResponse{
			Id:   categoryBrand.Brands.ID,
			Name: categoryBrand.Brands.Name,
			Logo: categoryBrand.Brands.Logo,
		})
	}

	brandListResponse.Data = brandInfoResponses

	return &brandListResponse, nil
}

func (s *GoodsServer) CreateCategoryBrand(ctx context.Context, req *proto.CategoryBrandRequest) (*proto.CategoryBrandResponse, error) {
	var category models.Category
	if result := global.DB.First(&category, req.CategoryId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "分类不存在")
	}
	var brand models.Brands
	if result := global.DB.First(&brand, req.BrandId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "品牌不存在")
	}

	categoryBrand := models.GoodsCategoryBrand{
		CategoryID: req.CategoryId,
		BrandsID:   req.BrandId,
	}

	global.DB.Save(&categoryBrand)
	return &proto.CategoryBrandResponse{Id: categoryBrand.ID}, nil
}

func (s *GoodsServer) DeleteCategoryBrand(ctx context.Context, req *proto.CategoryBrandRequest) (*emptypb.Empty, error) {
	if result := global.DB.Delete(&models.GoodsCategoryBrand{}, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "品牌分类不存在")
	}
	return &emptypb.Empty{}, nil
}
func (s *GoodsServer) UpdateCategoryBrand(ctx context.Context, req *proto.CategoryBrandRequest) (*emptypb.Empty, error) {
	var categoryBrand models.GoodsCategoryBrand
	if result := global.DB.First(&categoryBrand, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "品牌分类不存在")
	}
	var category models.Category
	if result := global.DB.First(&category, req.CategoryId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "分类不存在")
	}
	var brand models.Brands
	if result := global.DB.First(&brand, req.BrandId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "品牌不存在")
	}

	categoryBrand.CategoryID = req.CategoryId
	categoryBrand.BrandsID = req.BrandId
	global.DB.Save(&categoryBrand)
	return &emptypb.Empty{}, nil
}
