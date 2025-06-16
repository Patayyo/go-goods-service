package dto

type CreateGoodInput struct {
	Name string `json:"name"`
}

type UpdateGoodInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
