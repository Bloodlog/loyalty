package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLuhnCheck(t *testing.T) {
	assert.True(t, LuhnCheck(1234567890123456))
	assert.True(t, LuhnCheck(1234567812345678901))

	assert.False(t, LuhnCheck(12345))
}
