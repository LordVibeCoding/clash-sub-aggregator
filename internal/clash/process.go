package clash

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"clash-sub-aggregator/internal/model"
	"clash-sub-aggregator/internal/subscription"
)

type Process struct {
	cfg     model.MihomoConfig
	subMgr  *subscription.Manager
	cmd     *exec.Cmd
	mu      sync.Mutex
	running bool
	stopCh  chan struct{} // 用于通知监控 goroutine 是主动停止
}

func NewProcess(cfg model.MihomoConfig, subMgr *subscription.Manager) *Process {
	return &Process{
		cfg:    cfg,
		subMgr: subMgr,
	}
}

// Start 生成配置并启动 mihomo
func (p *Process) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return fmt.Errorf("mihomo 已在运行")
	}

	if err := p.writeConfig(); err != nil {
		return fmt.Errorf("生成配置失败: %w", err)
	}

	configPath := filepath.Join(p.cfg.ConfigDir, "config.yaml")
	p.cmd = exec.Command(p.cfg.Binary, "-f", configPath)
	p.cmd.Stdout = os.Stdout
	p.cmd.Stderr = os.Stderr

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("启动 mihomo 失败: %w", err)
	}

	p.running = true
	p.stopCh = make(chan struct{})
	log.Printf("mihomo 已启动, PID: %d", p.cmd.Process.Pid)

	// 监控进程退出
	stopCh := p.stopCh
	cmd := p.cmd
	go func() {
		err := cmd.Wait()
		select {
		case <-stopCh:
			// 主动停止，不修改状态
			return
		default:
		}
		p.mu.Lock()
		p.running = false
		p.mu.Unlock()
		if err != nil {
			log.Printf("mihomo 异常退出: %v", err)
		}
	}()

	return nil
}

// Stop 停止 mihomo
func (p *Process) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running || p.cmd == nil || p.cmd.Process == nil {
		return nil
	}

	// 通知监控 goroutine 这是主动停止
	close(p.stopCh)

	if err := p.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("停止 mihomo 失败: %w", err)
	}
	p.running = false
	p.cmd = nil
	log.Println("mihomo 已停止")

	// 等一下确保端口释放
	time.Sleep(500 * time.Millisecond)
	return nil
}

// Restart 重启 mihomo（重新生成配置）
func (p *Process) Restart() error {
	_ = p.Stop()
	return p.Start()
}

func (p *Process) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}

// ControllerAddr 返回 mihomo external controller 地址
func (p *Process) ControllerAddr() string {
	return fmt.Sprintf("http://127.0.0.1:%d", p.cfg.ControllerPort)
}

func (p *Process) writeConfig() error {
	proxies := p.subMgr.AllProxies()
	data, err := GenerateConfig(proxies, p.cfg)
	if err != nil {
		return err
	}
	configPath := filepath.Join(p.cfg.ConfigDir, "config.yaml")
	return os.WriteFile(configPath, data, 0644)
}
