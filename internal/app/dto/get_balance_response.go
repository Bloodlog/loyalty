package dto

type BalanceResponseBody struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
