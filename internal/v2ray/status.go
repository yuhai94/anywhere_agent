package v2ray

import (
	"os/exec"
	"strings"
)

// IsV2RayRunning 检查V2Ray是否正在运行
func IsV2RayRunning() bool {
	// 尝试使用systemctl检查
	cmd := exec.Command("systemctl", "is-active", "v2ray")
	output, err := cmd.CombinedOutput()
	if err == nil && strings.TrimSpace(string(output)) == "active" {
		return true
	}

	// 尝试使用service命令检查
	cmd = exec.Command("service", "v2ray", "status")
	output, err = cmd.CombinedOutput()
	if err == nil && (strings.Contains(string(output), "running") || strings.Contains(string(output), "active")) {
		return true
	}

	// 尝试使用ps命令检查进程
	cmd = exec.Command("bash", "-c", "ps aux | grep v2ray | grep -v grep")
	output, err = cmd.CombinedOutput()
	if err == nil && len(output) > 0 {
		return true
	}

	return false
}

// GetV2RayStatus 获取V2Ray状态
func GetV2RayStatus() (*DeployStatus, error) {
	installed, version, err := CheckV2Ray()
	if err != nil {
		return nil, err
	}

	return &DeployStatus{
		Installed: installed,
		Running:   IsV2RayRunning(),
		Version:   version,
		Progress:  100,
		Message:   "V2Ray status checked",
	}, nil
}
