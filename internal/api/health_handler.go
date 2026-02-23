package api

import (
	"net/http"

	"clash-sub-aggregator/internal/health"
)

type HealthHandler struct {
	checker *health.Checker
}

func NewHealthHandler(checker *health.Checker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

// Status 查看健康检查状态
func (h *HealthHandler) Status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.checker.Status())
}

// TriggerCheck 手动触发健康检查
func (h *HealthHandler) TriggerCheck(w http.ResponseWriter, r *http.Request) {
	go h.checker.CheckAll()
	writeJSON(w, http.StatusOK, map[string]any{"message": "健康检查已触发"})
}
