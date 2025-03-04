package dto

type WithdrawBody struct {
	OrderNumber string  `json:"order"`
	Sum         float64 `json:"sum"`
}
