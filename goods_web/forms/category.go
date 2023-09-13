package forms

type CategoryForm struct {
	Name           string `json:"name,omitempty" binding:"required"`
	Level          int32  `json:"level,omitempty" binding:"required,oneof=1 2 3"`
	ParentCategory int32  `json:"parent_category,omitempty"`
	IsTab          *bool  `json:"is_tab,omitempty" binding:"required"`
}
