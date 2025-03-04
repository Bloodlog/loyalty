package utils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetGetUserID(t *testing.T) {
	ctx := context.Background()

	userID := 123
	ctx = SetUserID(ctx, userID)

	retrievedUserID, err := GetUserID(ctx)

	assert.NoError(t, err)
	assert.Equal(t, userID, retrievedUserID)
}

func TestGetUserID_NoUserID(t *testing.T) {
	ctx := context.Background()

	_, err := GetUserID(ctx)

	assert.Error(t, err)
	assert.Equal(t, "unauthorized", err.Error())
}
