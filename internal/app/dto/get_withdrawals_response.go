package dto

type WithdrawalsResponseBody struct {
	Number     string  `json:"order"`
	Withdrawaw float64 `json:"sum"`
	CreatedAt  string  `json:"processed_at"`
}
