package forms

type CategoryBrandForm struct {
	CategoryId int32 `json:"category_id,omitempty" binding:"required"`
	BrandId    int32 `json:"brand_id,omitempty" binding:"required"`
}
