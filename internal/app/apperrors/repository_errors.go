package apperrors

import "errors"

var ErrDuplicateOrderID = errors.New("duplicate order ID")
var ErrDuplicateOrderIDAnotherUserID = errors.New("duplicate order ID with another user")
var ErrDuplicateLogin = errors.New("duplicate login")

var ErrBalanceNotEnought = errors.New("balance Not Enought")
var ErrOrderNotFound = errors.New("order not found")
