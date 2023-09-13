package models

// 实际开发过程中，尽量设置为不为Null
// https://zhuanlan.zhihu.com/p/73997266
type Category struct {
	BaseModel
	Name             string    `gorm:"type:varchar(20);not null" json:"name"`
	ParentCategoryID int32     `json:"parent"`
	ParentCategory   *Category `json:"-"`
	//gorm用法 foreignKey表示外键，reference表示外键对应本体哪个键，此处为主键
	SubCategory []*Category `gorm:"foreignKey:ParentCategoryID;reference:ID" json:"sub_category"`
	// 用int32原因，proto中没有int，避免过多的类型转换
	Level int32 `gorm:"type:int;not null;default:1" json:"level"`
	IsTab bool  `gorm:"default:false;not null" json:"is_tab"`
}

type Brands struct {
	BaseModel
	Name string `gorm:"type:varchar(20);not null"`
	Logo string `gorm:"type:varchar(200);default:'';not null"`
}

// 在gorm中，可以直接用many2many自动帮你创建多对多表
// 这里使用手动创建
type GoodsCategoryBrand struct {
	BaseModel
	//外键1
	CategoryID int32 `gorm:"type:int;index:idx_category_brand,unique"`
	Category   Category

	//外键2
	BrandsID int32 `gorm:"type:int;index:idx_category_brand,unique"`
	Brands   Brands
}

// 自定义表名
func (GoodsCategoryBrand) TableName() string {
	return "goodscategorybrand"
}

type Goods struct {
	BaseModel

	CategoryID int32 `gorm:"type:int;not null"`
	Category   Category
	BrandsID   int32 `gorm:"type:int;not null"`
	Brands     Brands

	OnSale   bool `gorm:"default:false;not null"`
	ShipFree bool `gorm:"default:false;not null"`
	IsNew    bool `gorm:"default:false;not null"`
	IsHot    bool `gorm:"default:false;not null"`

	Name    string `gorm:"type:varchar(50);not null"`
	GoodsSn string `gorm:"type:varchar(50);not null"`
	//没做, 在web层做
	ClickNum int32 `gorm:"type:int;default:0;not null"`
	//没做，联动订单服务完成
	SoldNum int32 `gorm:"type:int;default:0;not null"`
	//没做，由web层访问另外的用户点赞服务获取
	FavNum int32 `gorm:"type:int;default:0;not null"`

	MarketPrice     float32  `gorm:"type:int;not null"`
	ShopPrice       float32  `gorm:"type:int;not null"`
	GoodsBrief      string   `gorm:"type:varchar(100);not null"`
	Images          GormList `gorm:"type:varchar(1000);not null"`
	DescImages      GormList `gorm:"type:varchar(1000);not null"`
	GoodsFrontImage string   `gorm:"type:varchar(200);not null"`
}
