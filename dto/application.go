package dto

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

type CreateApplicationInput struct {
	Id        string `json:"id" example:"iot-api-gateway"`
	Content   string `json:"content" example:"workflow yaml content"`
	RepoPath  string `json:"repo_path" example:""`
	Namespace string `json:"namespace" example:""`
	ValueFile string `json:"value_file" example:"values.yaml"`
}

func (in *CreateApplicationInput) Validate() error {
	return validation.ValidateStruct(&in,
		validation.Field(&in.Id, validation.Required, validation.Length(3, 100)),
		validation.Field(&in.Content, validation.Required, validation.Min(10)),
	)
}
