package forms

type CreateOrderForm struct {
	Address string `json:"address,omitempty" binding:"required"`
	Name    string `json:"name,omitempty" binding:"required"`
	Mobile  string `json:"mobile,omitempty" binding:"required"`
	Post    string `json:"post,omitempty" binding:"required"`
}
