package utils

import (
	"github.com/theplant/luhn"
)

func LuhnCheck(number int64) bool {
	return luhn.Valid(int(number))
}
