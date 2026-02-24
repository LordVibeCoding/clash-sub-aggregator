package api

import (
	"io/fs"
	"net/http"
	"os"
	"strings"

	"clash-sub-aggregator/internal/clash"
	"clash-sub-aggregator/internal/health"
	"clash-sub-aggregator/internal/subscription"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(token string, subMgr *subscription.Manager, proc *clash.Process, hc *health.Checker) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	subHandler := NewSubscriptionHandler(subMgr, proc)
	proxyHandler := NewProxyHandler(proc)
	healthHandler := NewHealthHandler(hc)

	// API 路由（需要 token 认证）
	r.Route("/api", func(r chi.Router) {
		r.Use(TokenAuth(token))

		r.Get("/subscriptions", subHandler.List)
		r.Post("/subscriptions", subHandler.Add)
		r.Delete("/subscriptions/{id}", subHandler.Delete)
		r.Post("/subscriptions/refresh", subHandler.RefreshAll)
		r.Post("/subscriptions/{id}/refresh", subHandler.RefreshOne)

		r.Get("/proxies", proxyHandler.List)
		r.Put("/proxies/{group}/{name}", proxyHandler.Switch)
		r.Get("/proxies/{name}/delay", proxyHandler.Delay)

		r.Get("/status", proxyHandler.Status)

		r.Get("/health", healthHandler.Status)
		r.Post("/health/check", healthHandler.TriggerCheck)
	})

	// 静态文件（前端管理面板，无需认证）
	staticDir := "./static"
	if _, err := os.Stat(staticDir); err == nil {
		fileServer := http.FileServer(http.Dir(staticDir))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			// 如果文件存在，直接返回
			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				path = "index.html"
			}
			if _, err := fs.Stat(os.DirFS(staticDir), path); err == nil {
				fileServer.ServeHTTP(w, r)
				return
			}
			// SPA fallback: 返回 index.html
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
		})
	}

	return r
}
