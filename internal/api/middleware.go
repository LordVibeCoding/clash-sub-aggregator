package api

import (
	"net/http"
	"strings"
)

// TokenAuth 简单的 token 认证中间件
func TokenAuth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			auth := r.Header.Get("Authorization")
			if auth == "" {
				auth = r.URL.Query().Get("token")
			} else {
				auth = strings.TrimPrefix(auth, "Bearer ")
			}

			if auth != token {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
