package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"clash-sub-aggregator/internal/clash"

	"github.com/go-chi/chi/v5"
)

type ProxyHandler struct {
	proc   *clash.Process
	client *http.Client
}

func NewProxyHandler(proc *clash.Process) *ProxyHandler {
	return &ProxyHandler{
		proc:   proc,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// List 列出所有代理节点（转发 mihomo API）
func (h *ProxyHandler) List(w http.ResponseWriter, r *http.Request) {
	if !h.proc.IsRunning() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "mihomo 未运行"})
		return
	}

	resp, err := h.client.Get(h.proc.ControllerAddr() + "/proxies")
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// Switch 切换代理节点
func (h *ProxyHandler) Switch(w http.ResponseWriter, r *http.Request) {
	if !h.proc.IsRunning() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "mihomo 未运行"})
		return
	}

	group := chi.URLParam(r, "group")
	name := chi.URLParam(r, "name")

	body := fmt.Sprintf(`{"name":"%s"}`, name)
	apiURL := fmt.Sprintf("%s/proxies/%s", h.proc.ControllerAddr(), url.PathEscape(group))

	req, err := http.NewRequest("PUT", apiURL, strings.NewReader(body))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		writeJSON(w, http.StatusOK, map[string]any{
			"message": fmt.Sprintf("已切换到 %s", name),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// Delay 测试节点延迟
func (h *ProxyHandler) Delay(w http.ResponseWriter, r *http.Request) {
	if !h.proc.IsRunning() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "mihomo 未运行"})
		return
	}

	name := chi.URLParam(r, "name")
	apiURL := fmt.Sprintf("%s/proxies/%s/delay?timeout=5000&url=http://www.gstatic.com/generate_204",
		h.proc.ControllerAddr(), url.PathEscape(name))

	resp, err := h.client.Get(apiURL)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	writeJSON(w, resp.StatusCode, result)
}

// Status 服务状态
func (h *ProxyHandler) Status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"mihomo_running": h.proc.IsRunning(),
		"controller":     h.proc.ControllerAddr(),
	})
}
