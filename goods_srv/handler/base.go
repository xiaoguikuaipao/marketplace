package handler

import "gorm.io/gorm"

// paginate 基于gorm实现查询分页
func paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = 1
		}
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}
		offset := (page - 1) * pageSize
		//即查询并返回table中从offset开始的Limit条数据
		return db.Offset(offset).Limit(pageSize)
	}
}
