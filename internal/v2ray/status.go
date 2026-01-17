package v2ray

import (
	"os/exec"
	"strings"

	"github.com/yuhai94/anywhere_agent/internal/logger"
	"go.uber.org/zap"
)

// IsV2RayRunning 检查V2Ray是否正在运行
func IsV2RayRunning() bool {
	logger.Debug("Checking if V2Ray is running")

	// 尝试使用systemctl检查
	cmd := exec.Command("systemctl", "is-active", "v2ray")
	logger.Debug("Executing command", zap.String("command", cmd.String()))
	output, err := cmd.CombinedOutput()
	logger.Debug("systemctl is-active v2ray output",
		zap.String("output", strings.TrimSpace(string(output))),
		zap.Error(err))
	if err == nil && strings.TrimSpace(string(output)) == "active" {
		logger.Debug("V2Ray is running (systemctl)")
		return true
	}

	// 尝试使用service命令检查
	cmd = exec.Command("service", "v2ray", "status")
	logger.Debug("Executing command", zap.String("command", cmd.String()))
	output, err = cmd.CombinedOutput()
	logger.Debug("service v2ray status output",
		zap.String("output", string(output)[:100]+"..."), // 只显示前100字符
		zap.Error(err))
	if err == nil && (strings.Contains(string(output), "running") || strings.Contains(string(output), "active")) {
		logger.Debug("V2Ray is running (service)")
		return true
	}

	// 尝试使用ps命令检查进程
	cmd = exec.Command("bash", "-c", "ps aux | grep v2ray | grep -v grep")
	logger.Debug("Executing command", zap.String("command", cmd.String()))
	output, err = cmd.CombinedOutput()
	logger.Debug("ps command output",
		zap.String("output", string(output)),
		zap.Error(err))
	if err == nil && len(output) > 0 {
		logger.Debug("V2Ray is running (ps)")
		return true
	}

	logger.Debug("V2Ray is not running")
	return false
}

// GetV2RayStatus 获取V2Ray状态
func GetV2RayStatus() (*DeployStatus, error) {
	logger.Info("Getting V2Ray status")

	installed, version, err := CheckV2Ray()
	if err != nil {
		logger.Error("Failed to get V2Ray status", zap.Error(err))
		return nil, err
	}

	running := IsV2RayRunning()
	status := &DeployStatus{
		Installed: installed,
		Running:   running,
		Version:   version,
		Progress:  100,
		Message:   "V2Ray status checked",
	}

	logger.Info("V2Ray status retrieved",
		zap.Bool("installed", installed),
		zap.Bool("running", running),
		zap.String("version", version))

	return status, nil
}
