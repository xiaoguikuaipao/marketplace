package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type GormList []string

func (g GormList) Value() (driver.Value, error) {
	return json.Marshal(g)
}

func (g *GormList) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), g)
}

type BaseModel struct {
	ID        int32          `gorm:"primarykey;type:int" json:"id"`
	CreatedAt time.Time      `gorm:"created_at" json:"-"`
	UpdatedAt time.Time      `gorm:"updated_at" json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"deleted_at" json:"-"`
	//上述是软删除
	IsDeleted bool `gorm:"is_deleted" json:"-"`
}
