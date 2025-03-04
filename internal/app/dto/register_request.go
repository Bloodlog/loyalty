package dto

type RegisterRequestBody struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
