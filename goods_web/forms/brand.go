package forms

type BrandForm struct {
	Name string `json:"name,omitempty" binding:"required"`
	Logo string `json:"logo,omitempty" binding:"required"`
	Id   string `json:"id,omitempty"`
}

type BrandUpdateForm struct {
	Name string `json:"name,omitempty" binding:"required"`
	Logo string `json:"logo,omitempty" binding:"required"`
	Id   string `json:"id,omitempty" binding:"required"`
}
