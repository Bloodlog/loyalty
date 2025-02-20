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

const cookieName = "session_id"

//// Создаем экземпляр securecookie с секретными ключами
//var s = securecookie.New(
//	securecookie.GenerateRandomKey(32), // Key for authentication (HMAC)
//	securecookie.GenerateRandomKey(32), // Key for encryption
//)
//
//func (h *UserHandler) checkAndSetCookie(response http.ResponseWriter, request *http.Request, userID string) {
//	var sessionID string
//
//	// Проверяем существующую куку
//	cookie, err := request.Cookie(cookieName)
//	if err == nil {
//		err = s.Decode(cookieName, cookie.Value, &sessionID)
//		if err == nil {
//			h.Logger.Infof("Valid session cookie found: %s", sessionID)
//			return // Кука валидна, ничего не делаем
//		}
//	}
//
//	// Если куки нет или она повреждена, создаем новую
//	encoded, err := s.Encode(cookieName, userID)
//	if err != nil {
//		h.Logger.Error("Failed to encode session cookie", err)
//		return
//	}
//
//	http.SetCookie(response, &http.Cookie{
//		Name:     cookieName,
//		Value:    encoded,
//		Path:     "/",
//		HttpOnly: true,
//		Secure:   true,  // Включить в проде, если HTTPS
//		SameSite: http.SameSiteLaxMode,
//	})
//
//	h.Logger.Infof("New session cookie set for user: %s", userID)
//}
