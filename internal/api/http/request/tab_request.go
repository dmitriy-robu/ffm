package request

type TabRequest struct {
	Name        string `json:"name" validate:"required,max=255,min=2"`
	Description string `json:"description" validate:"required,max=10000"`
}

type TabUpdateRequest struct {
	Name        string `json:"name" validate:"required,max=255,min=2"`
	Description string `json:"description" validate:"required,max=10000"`
}
