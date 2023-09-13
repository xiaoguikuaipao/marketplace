package handler

import (
	"context"
	"fmt"

	"grpc/goods_srv/global"
	"grpc/goods_srv/models"
	"grpc/goods_srv/proto"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GoodsServer struct {
	*proto.UnimplementedGoodsServer
}

func ModelToResponse(goods models.Goods) *proto.GoodsInfoResponse {
	goodsInfoResponse := proto.GoodsInfoResponse{
		Id:              goods.ID,
		CategoryId:      goods.Category.ID,
		Name:            goods.Name,
		GoodsSn:         goods.GoodsSn,
		ClickNum:        goods.ClickNum,
		SoldNum:         goods.SoldNum,
		FavNum:          goods.FavNum,
		MarketPrice:     goods.MarketPrice,
		ShopPrice:       goods.ShopPrice,
		GoodsBrief:      goods.GoodsBrief,
		ShipFree:        goods.ShipFree,
		GoodsFrontImage: goods.GoodsFrontImage,
		IsNew:           goods.IsNew,
		IsHot:           goods.IsHot,
		OnSale:          goods.OnSale,
		DescImages:      goods.DescImages,
		Images:          goods.Images,
		Category: &proto.CategoryBriefInfoResponse{
			Id:   goods.Category.ID,
			Name: goods.Category.Name,
		},
		Brand: &proto.BrandInfoResponse{
			Id:   goods.Brands.ID,
			Name: goods.Brands.Name,
			Logo: goods.Brands.Logo,
		},
	}
	return &goodsInfoResponse
}

func (s *GoodsServer) GoodsList(c context.Context, req *proto.GoodsFilterRequest) (*proto.GoodsListResponse, error) {
	//关键词搜索，查询新品，查询热门商品，通过价格区间筛选，通过商品分类筛选
	goodsListResponse := &proto.GoodsListResponse{}
	var goods []models.Goods
	// 指明之后查询都是在goods表中进行
	localDB := global.DB.Model(&models.Goods{})
	// 使用localDB实例进行链式的条件递进筛选, 每一个条件用where过滤一次
	if req.KeyWords != "" {
		localDB = localDB.Where("name LIKE ?", "%"+req.KeyWords+"%")
	}
	if req.IsHot {
		localDB = localDB.Where(&models.Goods{IsHot: true})
	}
	if req.IsNew {
		localDB = localDB.Where(&models.Goods{IsNew: true})
	}
	if req.PriceMin > 0 {
		localDB = localDB.Where("shop_price >= ?", req.PriceMin)
	}
	if req.PriceMax > 0 {
		localDB = localDB.Where("shop_price <= ?", req.PriceMax)
	}
	if req.Brand > 0 {
		localDB = localDB.Where("brand_id=?", req.Brand)
	}

	//通过category去查询商品
	if req.TopCategory > 0 {
		var category models.Category
		if result := global.DB.First(&category, req.TopCategory); result.RowsAffected == 0 {
			return nil, status.Errorf(codes.NotFound, "分类不存在")
		}

		var subQuery string
		if category.Level == 1 {
			subQuery = fmt.Sprintf("SELECT id FROM category WHERE parent_category_id IN (SELECT id from category WHERE parent_category_id=%d)", req.TopCategory)
		} else if category.Level == 2 {
			subQuery = fmt.Sprintf("SELECT id FROM category WHERE parent_category_id=%d", req.TopCategory)
		} else if category.Level == 3 {
			subQuery = fmt.Sprintf("SELECT id FROM category WHERE id=%d", req.TopCategory)
		}
		localDB = localDB.Where(fmt.Sprintf("category_id in (%s)", subQuery))
	}
	var count int64
	localDB.Count(&count)
	goodsListResponse.Total = int32(count)

	result := localDB.Preload("Category").Preload("Brands").Scopes(paginate(int(req.Pages), int(req.PagePerNums))).Find(&goods)
	if result.Error != nil {
		return nil, result.Error
	}
	for _, good := range goods {
		goodsInfoResponse := ModelToResponse(good)
		goodsListResponse.Data = append(goodsListResponse.Data, goodsInfoResponse)
	}
	return goodsListResponse, nil
}

func (s *GoodsServer) BatchGetGoods(c context.Context, req *proto.BatchGoodsIdInfo) (*proto.GoodsListResponse, error) {
	var goods []models.Goods
	goodsListResponse := &proto.GoodsListResponse{}
	//这里传参req.Id必须是主键
	//否则应该用sql语句"xxx IN ?, xxx"
	if result := global.DB.Where(req.Id).Find(&goods); result.Error != nil {
		zap.L().Error("批量搜索商品失败", zap.Error(result.Error))
	} else {
		for _, good := range goods {
			goodsInfoResponse := ModelToResponse(good)
			goodsListResponse.Data = append(goodsListResponse.Data, goodsInfoResponse)
		}
		goodsListResponse.Total = int32(result.RowsAffected)
	}
	return goodsListResponse, nil
}

func (s *GoodsServer) CreateGoods(c context.Context, req *proto.CreateGoodsInfo) (*proto.GoodsInfoResponse, error) {
	//先明确创建的商品是什么类目，品牌
	//判断类目是否存在, 并保存类目信息
	var category models.Category
	if result := global.DB.First(&category, req.CategoryId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "分类不存在")
	}
	//判断品牌是否存在, 并保存品牌信息
	var brand models.Brands
	if result := global.DB.First(&brand, req.BrandId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "品牌不存在")
	}

	// redis 对token过滤： 防止同一页面的多次提交，或者页面信息过期
	//TODO

	good := models.Goods{
		CategoryID:  category.ID,
		Category:    category,
		BrandsID:    brand.ID,
		Brands:      brand,
		OnSale:      req.OnSale,
		ShipFree:    req.ShipFree,
		IsNew:       req.IsNew,
		IsHot:       req.IsHot,
		Name:        req.Name,
		GoodsSn:     req.GoodsSn,
		MarketPrice: req.MarketPrice,
		ShopPrice:   req.ShopPrice,
		GoodsBrief:  req.GoodsBrief,
		//这里的images是上传成功后返回的访问url
		Images:          req.Images,
		DescImages:      req.DescImages,
		GoodsFrontImage: req.GoodsFrontImage,
	}

	global.DB.Save(&good)
	return &proto.GoodsInfoResponse{
		Id: good.ID,
	}, nil
}
func (s *GoodsServer) DeleteGoods(c context.Context, req *proto.DeleteGoodsInfo) (*emptypb.Empty, error) {
	if result := global.DB.Delete(&models.Goods{}, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品不存在")
	}
	return &emptypb.Empty{}, nil
}
func (s *GoodsServer) UpdateGoods(c context.Context, req *proto.CreateGoodsInfo) (*emptypb.Empty, error) {
	var good models.Goods
	if result := global.DB.First(&good, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品不存在")
	}
	var category models.Category
	if result := global.DB.First(&category, req.CategoryId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "分类不存在")
	}
	var brand models.Brands
	if result := global.DB.First(&brand, req.BrandId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "品牌不存在")
	}
	good.GoodsSn = req.GoodsSn
	good.GoodsBrief = req.GoodsBrief
	good.GoodsFrontImage = req.GoodsFrontImage
	good.Brands = brand
	good.BrandsID = req.BrandId
	good.CategoryID = req.CategoryId
	good.Category = category
	good.Name = req.Name
	good.DescImages = req.DescImages
	good.Images = req.Images
	good.ShopPrice = req.ShopPrice
	good.MarketPrice = req.MarketPrice
	good.IsHot = req.IsHot
	good.IsNew = req.IsNew
	good.ShipFree = req.ShipFree
	good.OnSale = req.OnSale
	global.DB.Save(&good)
	return &emptypb.Empty{}, nil
}

func (s *GoodsServer) GetGoodsDetail(c context.Context, req *proto.GoodInfoRequest) (*proto.GoodsInfoResponse, error) {
	var good models.Goods

	if result := global.DB.Preload("Category").Preload("Brands").First(&good, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品不存在")
	}
	goodsInfoResponse := ModelToResponse(good)
	return goodsInfoResponse, nil
}
