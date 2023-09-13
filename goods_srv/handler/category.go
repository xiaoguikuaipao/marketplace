package handler

import (
	"context"
	"encoding/json"
	"time"

	"grpc/goods_srv/global"
	"grpc/goods_srv/models"
	"grpc/goods_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// 商品分类
func (s *GoodsServer) GetAllCategorysList(c context.Context, req *emptypb.Empty) (*proto.CategoryListResponse, error) {
	/*
		[
			{
				"id" : xx,
				"name":xx,
				"level":1
				"sub_category":[
						"id":xx,
						"name":xx,
						"level":2,
						"sub_category":[

						]
				], [  xxx  ]
			},
			{xxx},
		]
	*/
	var categorys []models.Category
	// 注意此处用法,
	//1. 首先用where过滤出1级类目
	//2. 然后预加载二级类目和三级类目Sub.Sub
	global.DB.Where(&models.Category{Level: 1}).Preload("SubCategory.SubCategory").Find(&categorys)
	b, _ := json.Marshal(&categorys)
	return &proto.CategoryListResponse{JsonData: string(b)}, nil
}

func (s *GoodsServer) GetSubCategory(c context.Context, req *proto.CategoryListRequest) (*proto.SubCategoryListResponse, error) {
	var rsp proto.SubCategoryListResponse
	var category models.Category
	if result := global.DB.First(&category, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品分类不存在")
	}

	// 先填充他的类目信息
	rsp.Info = &proto.CategoryInfoResponse{
		Id:             category.ID,
		Name:           category.Name,
		ParentCategory: category.ParentCategoryID,
		Level:          category.ID,
		IsTab:          category.IsTab,
	}

	//再获取并填充该类目的子类目信息
	var subCategorys []models.Category
	var subCategoryListResponse []*proto.CategoryInfoResponse
	//无需递归和预加载
	//preloads := "SubCategory"
	//if category.Level == 1 {
	//	preloads = "SubCategory.SubCategory"
	//}
	global.DB.Where(&models.Category{ParentCategoryID: req.Id}).Find(&subCategorys)
	for _, subCategory := range subCategorys {
		subCategoryListResponse = append(subCategoryListResponse, &proto.CategoryInfoResponse{
			Id:             subCategory.ID,
			Name:           subCategory.Name,
			ParentCategory: subCategory.ParentCategoryID,
			Level:          subCategory.Level,
			IsTab:          subCategory.IsTab,
		})
	}
	rsp.SubCategorys = subCategoryListResponse
	return &rsp, nil
}

func (s *GoodsServer) CreateCategory(c context.Context, req *proto.CategoryInfoRequest) (*proto.CategoryInfoResponse, error) {
	cMap := map[string]interface{}{}
	cMap["name"] = req.Name
	cMap["is_tab"] = req.IsTab
	cMap["level"] = req.Level
	if req.Level != 1 {
		//去查询父类目是否存在
		cMap["parent_category_id"] = req.ParentCategory
	}
	global.DB.Model(&models.Category{}).Create(cMap)
	return &proto.CategoryInfoResponse{Name: req.Name}, nil
}

func (s *GoodsServer) DeleteCategory(c context.Context, req *proto.DeleteCategoryRequest) (*emptypb.Empty, error) {
	if result := global.DB.Delete(&models.Category{}, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品分类不存在")
	}
	return &emptypb.Empty{}, nil
}

func (s *GoodsServer) UpdateCategory(c context.Context, req *proto.CategoryInfoRequest) (*emptypb.Empty, error) {
	var category models.Category

	if result := global.DB.First(&category, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品分类不存在")
	}

	if req.Name != "" {
		category.Name = req.Name
	}
	if req.ParentCategory != 0 {
		category.ParentCategoryID = req.ParentCategory
	}
	if req.Level != 0 {
		category.Level = req.Level
	}
	category.IsTab = req.IsTab
	category.UpdatedAt = time.Now()

	global.DB.Save(&category)
	return &emptypb.Empty{}, nil
}
