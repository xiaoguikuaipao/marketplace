package forms

// 用于创建或者全量更新
type GoodsForm struct {
	Name        string   `form:"name" json:"name,omitempty" binding:"required"`
	GoodsSn     string   `form:"goods_sn" json:"goods_sn,omitempty" binding:"required"`
	Stocks      int32    `form:"stocks" json:"stocks,omitempty" binding:"required"`
	CategoryId  int32    `form:"category_id" json:"category_id,omitempty" binding:"required"`
	MarketPrice float32  `form:"market_price" json:"market_price,omitempty" binding:"required"`
	ShopPrice   float32  `form:"shop_price" json:"shop_price,omitempty" binding:"required"`
	GoodsBrief  string   `form:"goods_brief" json:"goods_brief,omitempty"`
	Images      []string `form:"images" json:"images,omitempty"`
	DescImages  []string `form:"desc_images" json:"desc_images,omitempty"`
	//注意，表单验证时应设置为指针，否则默认值false会被认为没有传
	ShipFree   *bool  `form:"ship_free" json:"ship_free,omitempty" binding:"required"`
	FrontImage string `form:"front_image" json:"front_image,omitempty"`
	Brand      int32  `form:"brand" json:"brand,omitempty" binding:"required"`
}

// 用于部分更新
type GoodsStatusForm struct {
	IsNew  *bool `form:"new" binding:"required" json:"is_new"`
	IsHot  *bool `form:"hot" binding:"required" json:"is_hot"`
	OnSale *bool `form:"sale" binding:"required" json:"on_sale"`
}
