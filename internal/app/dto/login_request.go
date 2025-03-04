package dto

type LoginRequestBody struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
