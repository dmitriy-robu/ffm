package request

import "mime/multipart"

type GoalRequest struct {
	File multipart.FileHeader `json:"file" validate:"required"`
	//Name string               `json:"name" validate:"required,max=255,min=2"`
}

type GoalUpdateRequest struct {
	File *multipart.FileHeader `json:"file,omitempty" validate:"omitempty"`
	Name string                `json:"name" validate:"required,max=255,min=2"`
}
