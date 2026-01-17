package v2ray

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/yuhai94/anywhere_agent/internal/logger"
	"go.uber.org/zap"
)

// DeployStatus V2Ray部署状态
type DeployStatus struct {
	Installed bool   `json:"installed"`
	Running   bool   `json:"running"`
	Version   string `json:"version"`
	Progress  int    `json:"progress"`
	Message   string `json:"message"`
}

// CheckV2Ray 检查V2Ray是否已安装
func CheckV2Ray() (bool, string, error) {
	// 1. 使用ps命令检查v2ray进程是否存在
	psCmd := exec.Command("bash", "-c", "ps aux | grep -v grep | grep v2ray")
	psOutput, err := psCmd.CombinedOutput()
	logger.Debug("ps command output", zap.String("ps_output", string(psOutput)))
	if err != nil {
		// 检查是否存在v2ray进程
		if strings.Contains(err.Error(), "exit status 1") {
			// grep没有找到匹配，即没有v2ray进程
			return false, "", nil
		}
		return false, "", fmt.Errorf("failed to check v2ray process: %w", err)
	}

	// 2. 如果有进程，尝试获取版本信息
	var version string
	versionCmd := exec.Command("v2ray", "--version")
	versionOutput, err := versionCmd.CombinedOutput()
	if err == nil {
		version = strings.TrimSpace(string(versionOutput))
	} else {
		version = "unknown"
	}

	// 3. 检查是否正在运行
	isRunning := IsV2RayRunning()

	return isRunning, version, nil
}

// DeployV2Ray 部署V2Ray
func DeployV2Ray(port int, uuid string, accessLogPath string) (*DeployStatus, error) {
	status := &DeployStatus{
		Progress: 0,
		Message:  "Starting V2Ray deployment",
	}

	// 检查是否已安装
	installed, version, err := CheckV2Ray()
	if err != nil {
		return status, err
	}

	if installed {
		status.Installed = true
		status.Version = version
		status.Progress = 100
		status.Message = "V2Ray already installed"
		return status, nil
	}

	// 1. 直接执行V2Ray安装脚本
	status.Progress = 20
	status.Message = "Downloading and installing V2Ray"

	// 使用curl直接执行脚本，不保存到本地
	installCmd := exec.Command("bash", "-c", "curl -L https://github.com/v2fly/fhs-install-v2ray/raw/master/install-release.sh | bash -s -- --force")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		return status, fmt.Errorf("failed to install v2ray: %w", err)
	}

	status.Progress = 40
	status.Message = "V2Ray installation completed"

	// 3. 配置V2Ray
	status.Progress = 60
	status.Message = "Configuring V2Ray"

	if err := configureV2Ray(port, uuid, accessLogPath); err != nil {
		return status, fmt.Errorf("failed to configure v2ray: %w", err)
	}

	// 4. 启动V2Ray服务
	status.Progress = 80
	status.Message = "Starting V2Ray service"

	startCmd := exec.Command("systemctl", "start", "v2ray")
	if err := startCmd.Run(); err != nil {
		// 尝试使用service命令
		startCmd = exec.Command("service", "v2ray", "start")
		if err := startCmd.Run(); err != nil {
			return status, fmt.Errorf("failed to start v2ray: %w", err)
		}
	}

	// 5. 设置开机自启
	enableCmd := exec.Command("systemctl", "enable", "v2ray")
	if err := enableCmd.Run(); err != nil {
		// 尝试使用chkconfig
		enableCmd = exec.Command("chkconfig", "v2ray", "on")
		if err != nil {
			// 忽略开机自启错误，不影响主功能
			logger.Warn("Failed to enable v2ray service", zap.Error(err))
		}
	}

	// 6. 验证安装
	status.Progress = 100
	status.Message = "V2Ray deployment completed"

	installed, version, err = CheckV2Ray()
	if err != nil {
		return status, err
	}

	status.Installed = installed
	status.Version = version
	status.Running = IsV2RayRunning()

	return status, nil
}

// configureV2Ray 配置V2Ray
func configureV2Ray(port int, uuid string, accessLogPath string) error {
	// V2Ray配置文件路径
	configPath := "/usr/local/etc/v2ray/config.json"

	// 读取现有配置
	existingConfig, err := os.ReadFile(configPath)
	if err != nil {
		// 如果配置文件不存在，创建新的
		if os.IsNotExist(err) {
			return createNewConfig(configPath, port, uuid, accessLogPath)
		}
		return fmt.Errorf("failed to read v2ray config: %w", err)
	}

	// 检查配置是否已包含我们需要的配置
	if strings.Contains(string(existingConfig), uuid) && strings.Contains(string(existingConfig), fmt.Sprintf("%d", port)) {
		return nil // 配置已存在，无需修改
	}

	// 创建新配置
	return createNewConfig(configPath, port, uuid, accessLogPath)
}

// createNewConfig 创建新的V2Ray配置文件
func createNewConfig(configPath string, port int, uuid string, accessLogPath string) error {
	// V2Ray配置模板
	configTemplate := fmt.Sprintf(`{
  "log": {
    "access": "%s",
    "error": "/var/log/v2ray/error.log",
    "loglevel": "info"
  },
  "inbounds": [
    {
      "port": %d,
      "tag": "vmess",
      "protocol": "vmess",
      "settings": {
        "clients": [
          {
            "email": "default",
            "id": "%s",
            "alterId": 0
          }
        ]
      }
    }
  ],
  "outbounds": [
    {
      "protocol": "freedom",
      "tag": "direct",
      "settings": {}
    }
  ]
}`, accessLogPath, port, uuid)

	// 确保配置目录存在
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 写入配置文件
	if err := os.WriteFile(configPath, []byte(configTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write v2ray config: %w", err)
	}

	// 确保日志目录存在
	logDir := "/var/log/v2ray"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	return nil
}
