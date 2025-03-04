package dto

type WithdrawalsResponseBody struct {
	Number     string  `json:"order"`
	CreatedAt  string  `json:"processed_at"`
	Withdrawaw float64 `json:"sum"`
}
