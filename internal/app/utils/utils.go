package utils

import (
	"strconv"
)

func LuhnCheck(number int64) bool {
	numStr := strconv.FormatInt(number, 10)
	length := len(numStr)

	return length >= 8 && length <= 19
}
