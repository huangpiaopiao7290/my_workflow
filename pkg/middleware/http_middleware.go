package middleware

import (
	"context"
	"my_workflow/pkg/common/helper"
	"net/http"
)

// http中间件： 仅做适配grpc拦截器验证
func HttpAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "missing authorization", http.StatusUnauthorized)
			return
		}
		
		userID, err := helper.ValidateToken(token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		
		r = r.WithContext(context.WithValue(r.Context(), "user_id", userID))
		next.ServeHTTP(w, r)
	})
}