package middlewares

import (
	"loyalty/internal/app/repositories"
	"loyalty/internal/app/services"
	"loyalty/internal/app/utils"
	"net/http"
)

func Auth(
	jwtService services.JwtService,
	userRepository repositories.UserRepositoryInterface,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "", http.StatusUnauthorized)
				return
			}

			userID, err := jwtService.GetUserID(token)
			if err != nil {
				http.Error(w, "", http.StatusUnauthorized)
				return
			}
			ctx := utils.SetUserID(r.Context(), userID)

			exist := userRepository.IsExistByID(ctx, userID)
			if !exist {
				http.Error(w, "", http.StatusUnauthorized)
				return
			}

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
