package request

type WorkoutRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type VideoSavePositionRequest struct {
	Position float64 `json:"position" validate:"required"`
}

type VideoUpdateRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}
