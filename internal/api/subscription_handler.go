package api

import (
	"encoding/json"
	"net/http"

	"clash-sub-aggregator/internal/clash"
	"clash-sub-aggregator/internal/subscription"

	"github.com/go-chi/chi/v5"
)

type SubscriptionHandler struct {
	mgr  *subscription.Manager
	proc *clash.Process
}

func NewSubscriptionHandler(mgr *subscription.Manager, proc *clash.Process) *SubscriptionHandler {
	return &SubscriptionHandler{mgr: mgr, proc: proc}
}

func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	subs := h.mgr.List()
	// 返回时隐藏代理详情，只返回数量
	type subInfo struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		URL        string `json:"url"`
		ProxyCount int    `json:"proxy_count"`
		UpdatedAt  string `json:"updated_at"`
	}
	var result []subInfo
	for _, s := range subs {
		result = append(result, subInfo{
			ID:         s.ID,
			Name:       s.Name,
			URL:        s.URL,
			ProxyCount: len(s.Proxies),
			UpdatedAt:  s.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"subscriptions": result})
}

type addRequest struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func (h *SubscriptionHandler) Add(w http.ResponseWriter, r *http.Request) {
	var req addRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求格式错误"})
		return
	}
	if req.Name == "" || req.URL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "name 和 url 不能为空"})
		return
	}

	sub, err := h.mgr.Add(req.Name, req.URL)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	// 添加成功后启动/重启 mihomo 加载新节点
	if h.proc.IsRunning() {
		_ = h.proc.Restart()
	} else {
		_ = h.proc.Start()
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message":     "订阅添加成功",
		"id":          sub.ID,
		"proxy_count": len(sub.Proxies),
	})
}

func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.mgr.Delete(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	if h.proc.IsRunning() {
		_ = h.proc.Restart()
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "已删除"})
}

func (h *SubscriptionHandler) RefreshAll(w http.ResponseWriter, r *http.Request) {
	total, err := h.mgr.RefreshAll()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error":       err.Error(),
			"proxy_count": total,
		})
		return
	}

	if h.proc.IsRunning() {
		_ = h.proc.Restart()
	} else {
		_ = h.proc.Start()
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message":     "刷新完成",
		"proxy_count": total,
	})
}

func (h *SubscriptionHandler) RefreshOne(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sub, err := h.mgr.RefreshOne(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	if h.proc.IsRunning() {
		_ = h.proc.Restart()
	} else {
		_ = h.proc.Start()
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message":     "刷新完成",
		"proxy_count": len(sub.Proxies),
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
