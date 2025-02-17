package dto

type BalanceResponseBody struct {
	Current   float64 `json:"current"`
	Withdrawn int64   `json:"withdrawn"`
}
