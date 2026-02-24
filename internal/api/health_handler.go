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

// Status 查看健康检查状态（含有效节点及延迟）
func (h *HealthHandler) Status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.checker.Status())
}

// TriggerCheck 触发健康检查，同步返回所有节点测速结果
func (h *HealthHandler) TriggerCheck(w http.ResponseWriter, r *http.Request) {
	results := h.checker.CheckAll()
	if results == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"message": "检查跳过（正在进行中或 mihomo 未运行）",
		})
		return
	}

	// 分离有效和无效节点
	var available []map[string]any
	var failed []string
	for _, r := range results {
		if r.Delay > 0 {
			available = append(available, map[string]any{
				"name":  r.Name,
				"delay": r.Delay,
			})
		} else {
			failed = append(failed, r.Name)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"total":           len(results),
		"available_count": len(available),
		"available":       available,
		"failed_count":    len(failed),
		"failed":          failed,
	})
}
