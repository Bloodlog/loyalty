package dto

type WithdrawBody struct {
	OrderNumber int64 `json:"order"`
	Sum         int64 `json:"sum"`
}
