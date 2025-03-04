package utils

import (
	"context"
	"errors"
)

type contextKey string

const userIDKey contextKey = "userID"

func SetUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func GetUserID(ctx context.Context) (int, error) {
	userID, ok := ctx.Value(userIDKey).(int)
	if !ok {
		return 0, errors.New("unauthorized")
	}
	return userID, nil
}
