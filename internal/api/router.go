package api

import (
	"clash-sub-aggregator/internal/clash"
	"clash-sub-aggregator/internal/subscription"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(token string, subMgr *subscription.Manager, proc *clash.Process) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(TokenAuth(token))

	subHandler := NewSubscriptionHandler(subMgr, proc)
	proxyHandler := NewProxyHandler(proc)

	r.Route("/api", func(r chi.Router) {
		// 订阅管理
		r.Get("/subscriptions", subHandler.List)
		r.Post("/subscriptions", subHandler.Add)
		r.Delete("/subscriptions/{id}", subHandler.Delete)
		r.Post("/subscriptions/refresh", subHandler.RefreshAll)
		r.Post("/subscriptions/{id}/refresh", subHandler.RefreshOne)

		// 代理控制
		r.Get("/proxies", proxyHandler.List)
		r.Put("/proxies/{group}/{name}", proxyHandler.Switch)
		r.Get("/proxies/{name}/delay", proxyHandler.Delay)

		// 状态
		r.Get("/status", proxyHandler.Status)
	})

	return r
}
