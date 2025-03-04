package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLuhnCheck(t *testing.T) {
	assert.True(t, LuhnCheck(2377225624))
	assert.True(t, LuhnCheck(24141463521))

	assert.False(t, LuhnCheck(12345))
}
