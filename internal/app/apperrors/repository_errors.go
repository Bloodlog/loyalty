package apperrors

import "errors"

var ErrDuplicateOrderID = errors.New("duplicate order ID")
var ErrDuplicateLogin = errors.New("duplicate login")

var ErrBalanceNotEnought = errors.New("balance Not Enought")
