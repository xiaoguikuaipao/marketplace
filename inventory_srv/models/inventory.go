package models

import (
	"database/sql/driver"
	"encoding/json"
)

// 实际开发过程中，尽量设置为不为Null
// https://zhuanlan.zhihu.com/p/73997266

type Inventory struct {
	BaseModel
	Goods  int32 `gorm:"type:int;index"`
	Stocks int32 `gorm:"type:int"`
	//分布式锁：乐观锁
	Version int32 `gorm:"type:int"`
}

type GoodsDetailList []GoodsDetail
type GoodsDetail struct {
	Goods int32
	Nums  int32
}

func (g GoodsDetailList) Value() (driver.Value, error) {
	return json.Marshal(g)
}

func (g *GoodsDetailList) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), g)
}

type StockRebackDetail struct {
	Status  int32           `gorm:"type:int"`
	OrderSn string          `gorm:"type:varchar(200);index:idx_order_sn,unique;"`
	Detail  GoodsDetailList `gorm:"type:varchar(200)"`
}

func (StockRebackDetail) TableName() string {
	return "stockrebackdetail"
}

//type InventoryHistory struct {
//	user int32
//	goods int32
//	nums int32
//	order int32
//	status int32
//}
