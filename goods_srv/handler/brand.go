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

// BrandList 根据请求返回品牌列表
func (s *GoodsServer) BrandList(c context.Context, req *proto.BrandFilterRequest) (*proto.BrandListResponse, error) {
	//1. 创建model结构体和返回结构体
	brandListResponse := proto.BrandListResponse{}
	var brandResponses []*proto.BrandInfoResponse
	var brands []models.Brands
	//2. 根据请求读DB，注意：又要根据分页情况存入model，又要得到总数
	result := global.DB.Scopes(paginate(int(req.Pages), int(req.PagePerNums))).Find(&brands)
	if result.Error != nil {
		return nil, result.Error
	}
	for _, brand := range brands {
		brandResponse := proto.BrandInfoResponse{
			Id:   brand.ID,
			Name: brand.Name,
			Logo: brand.Logo,
		}
		brandResponses = append(brandResponses, &brandResponse)
	}

	// 该方法用于获得table的总记录数
	var total int64
	global.DB.Model(&models.Brands{}).Count(&total)

	brandListResponse.Data = brandResponses
	brandListResponse.Total = int32(total)
	return &brandListResponse, nil
}

func (s *GoodsServer) CreateBrand(c context.Context, req *proto.BrandRequest) (*proto.BrandInfoResponse, error) {
	//先查询品牌是否存在
	if result := global.DB.Where("name=?", req.Name).First(&models.Brands{}); result.RowsAffected == 1 {
		return nil, status.Errorf(codes.InvalidArgument, "品牌已存在")
	}
	// 创建品牌
	brand := &models.Brands{
		Name: req.Name,
		Logo: req.Logo,
	}
	global.DB.Save(brand)

	return &proto.BrandInfoResponse{
		Id:   brand.ID,
		Name: brand.Name,
		Logo: brand.Logo,
	}, nil
}

func (s *GoodsServer) DeleteBrand(c context.Context, req *proto.BrandRequest) (*emptypb.Empty, error) {
	if result := global.DB.Delete(&models.Brands{}, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "品牌不存在")
	}
	return &emptypb.Empty{}, nil
}

func (s *GoodsServer) UpdateBrand(c context.Context, req *proto.BrandRequest) (*emptypb.Empty, error) {
	//先判断是否存在
	brands := models.Brands{}
	if result := global.DB.First(&brands); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "品牌不存在")
	}

	if req.Name != "" {
		brands.Name = req.Name
	}
	if req.Logo != "" {
		brands.Logo = req.Logo
	}

	global.DB.Save(&brands)

	return &emptypb.Empty{}, nil
}
