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
	logger.Debug("Executing command", zap.String("command", psCmd.String()))
	psOutput, err := psCmd.CombinedOutput()
	logger.Debug("ps command output",
		zap.String("output", string(psOutput)),
		zap.Error(err))
	if err != nil {
		// 检查是否存在v2ray进程
		if strings.Contains(err.Error(), "exit status 1") {
			// grep没有找到匹配，即没有v2ray进程
			logger.Info("No V2Ray process found")
			return false, "", nil
		}
		return false, "", fmt.Errorf("failed to check v2ray process: %w", err)
	}
	logger.Info("V2Ray process found")

	// 2. 如果有进程，尝试获取版本信息
	var version string
	versionCmd := exec.Command("v2ray", "--version")
	logger.Debug("Executing command", zap.String("command", versionCmd.String()))
	versionOutput, err := versionCmd.CombinedOutput()
	logger.Debug("v2ray version command output",
		zap.String("output", string(versionOutput)),
		zap.Error(err))
	if err == nil {
		version = strings.TrimSpace(string(versionOutput))
		logger.Info("V2Ray version detected", zap.String("version", version))
	} else {
		version = "unknown"
		logger.Warn("Failed to get V2Ray version", zap.Error(err))
	}

	// 3. 检查是否正在运行
	isRunning := IsV2RayRunning()
	logger.Info("V2Ray status check completed",
		zap.Bool("running", isRunning),
		zap.String("version", version))

	return isRunning, version, nil
}

// DeployV2Ray 部署V2Ray
func DeployV2Ray(port int, uuid string, accessLogPath string) (*DeployStatus, error) {
	// 创建一个默认的stopChan，不支持取消
	// 这个版本保留向后兼容，实际使用中应该调用带stopChan参数的版本
	stopChan := make(chan struct{})
	return DeployV2RayWithContext(port, uuid, accessLogPath, stopChan)
}

// DeployV2RayWithContext 部署V2Ray，支持通过stopChan取消部署
func DeployV2RayWithContext(port int, uuid string, accessLogPath string, stopChan <-chan struct{}) (*DeployStatus, error) {
	logger.Info("Starting V2Ray deployment",
		zap.Int("port", port),
		zap.String("uuid", uuid[:8]+"..."), // 只显示UUID前8位
		zap.String("access_log", accessLogPath))

	status := &DeployStatus{
		Progress: 0,
		Message:  "Starting V2Ray deployment",
	}

	// 检查是否已收到停止信号
	select {
	case <-stopChan:
		logger.Info("V2Ray deployment canceled due to stop signal")
		status.Message = "Deployment canceled"
		return status, nil
	default:
		// 继续执行
	}

	// 检查是否已安装
	installed, version, err := CheckV2Ray()
	if err != nil {
		logger.Error("Failed to check V2Ray installation", zap.Error(err))
		return status, err
	}

	// 检查是否已收到停止信号
	select {
	case <-stopChan:
		logger.Info("V2Ray deployment canceled due to stop signal")
		status.Message = "Deployment canceled"
		return status, nil
	default:
		// 继续执行
	}

	if installed {
		status.Installed = true
		status.Version = version
		status.Progress = 100
		status.Message = "V2Ray already installed"
		logger.Info("V2Ray already installed", zap.String("version", version))
		return status, nil
	}

	// 1. 直接执行V2Ray安装脚本
	status.Progress = 20
	status.Message = "Downloading and installing V2Ray"
	logger.Info("Downloading and installing V2Ray")

	// 检查是否已收到停止信号
	select {
	case <-stopChan:
		logger.Info("V2Ray deployment canceled due to stop signal")
		status.Message = "Deployment canceled"
		return status, nil
	default:
		// 继续执行
	}

	// 使用curl直接执行脚本，不保存到本地
	installCmd := exec.Command("bash", "-c", "curl -L https://github.com/v2fly/fhs-install-v2ray/raw/master/install-release.sh | bash -s -- --force")
	logger.Debug("Executing command", zap.String("command", installCmd.String()))
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	// 在goroutine中执行安装命令，支持取消
	installDone := make(chan error, 1)
	go func() {
		installDone <- installCmd.Run()
	}()

	// 等待安装完成或收到停止信号
	select {
	case <-stopChan:
		logger.Info("V2Ray installation canceled due to stop signal")
		// 尝试终止安装进程
		if installCmd.Process != nil {
			installCmd.Process.Kill()
			logger.Info("Killed V2Ray installation process")
		}
		status.Message = "Installation canceled"
		return status, nil
	case err := <-installDone:
		if err != nil {
			logger.Error("Failed to install V2Ray", zap.Error(err))
			return status, fmt.Errorf("failed to install v2ray: %w", err)
		}
	}

	logger.Info("V2Ray installation completed")

	// 检查是否已收到停止信号
	select {
	case <-stopChan:
		logger.Info("V2Ray deployment canceled due to stop signal")
		status.Message = "Deployment canceled"
		return status, nil
	default:
		// 继续执行
	}

	status.Progress = 40
	status.Message = "V2Ray installation completed"

	// 3. 配置V2Ray
	status.Progress = 60
	status.Message = "Configuring V2Ray"
	logger.Info("Configuring V2Ray")

	// 检查是否已收到停止信号
	select {
	case <-stopChan:
		logger.Info("V2Ray deployment canceled due to stop signal")
		status.Message = "Deployment canceled"
		return status, nil
	default:
		// 继续执行
	}

	if err := configureV2Ray(port, uuid, accessLogPath); err != nil {
		logger.Error("Failed to configure V2Ray", zap.Error(err))
		return status, fmt.Errorf("failed to configure v2ray: %w", err)
	}
	logger.Info("V2Ray configuration completed")

	// 检查是否已收到停止信号
	select {
	case <-stopChan:
		logger.Info("V2Ray deployment canceled due to stop signal")
		status.Message = "Deployment canceled"
		return status, nil
	default:
		// 继续执行
	}

	// 4. 启动V2Ray服务
	status.Progress = 80
	status.Message = "Starting V2Ray service"
	logger.Info("Starting V2Ray service")

	// 检查是否已收到停止信号
	select {
	case <-stopChan:
		logger.Info("V2Ray deployment canceled due to stop signal")
		status.Message = "Deployment canceled"
		return status, nil
	default:
		// 继续执行
	}

	startCmd := exec.Command("systemctl", "start", "v2ray")
	logger.Debug("Executing command", zap.String("command", startCmd.String()))
	startOutput, startErr := startCmd.CombinedOutput()
	logger.Debug("systemctl start v2ray output",
		zap.String("output", string(startOutput)),
		zap.Error(startErr))

	if startErr != nil {
		logger.Warn("Failed to start V2Ray with systemctl, trying service command", zap.Error(startErr))
		// 尝试使用service命令
		startCmd = exec.Command("service", "v2ray", "start")
		logger.Debug("Executing command", zap.String("command", startCmd.String()))
		startOutput, startErr = startCmd.CombinedOutput()
		logger.Debug("service start v2ray output",
			zap.String("output", string(startOutput)),
			zap.Error(startErr))
		if startErr != nil {
			logger.Error("Failed to start V2Ray service", zap.Error(startErr))
			return status, fmt.Errorf("failed to start v2ray: %w", startErr)
		}
	}
	logger.Info("V2Ray service started successfully")

	// 5. 设置开机自启
	logger.Info("Setting V2Ray to start on boot")
	enableCmd := exec.Command("systemctl", "enable", "v2ray")
	logger.Debug("Executing command", zap.String("command", enableCmd.String()))
	enableOutput, enableErr := enableCmd.CombinedOutput()
	logger.Debug("systemctl enable v2ray output",
		zap.String("output", string(enableOutput)),
		zap.Error(enableErr))

	if enableErr != nil {
		logger.Warn("Failed to enable V2Ray with systemctl, trying chkconfig", zap.Error(enableErr))
		// 尝试使用chkconfig
		enableCmd = exec.Command("chkconfig", "v2ray", "on")
		logger.Debug("Executing command", zap.String("command", enableCmd.String()))
		enableOutput, enableErr = enableCmd.CombinedOutput()
		logger.Debug("chkconfig enable v2ray output",
			zap.String("output", string(enableOutput)),
			zap.Error(enableErr))
		if enableErr != nil {
			// 忽略开机自启错误，不影响主功能
			logger.Warn("Failed to enable v2ray service on boot", zap.Error(enableErr))
		} else {
			logger.Info("V2Ray enabled on boot with chkconfig")
		}
	} else {
		logger.Info("V2Ray enabled on boot with systemctl")
	}

	// 6. 验证安装
	status.Progress = 100
	status.Message = "V2Ray deployment completed"
	logger.Info("Verifying V2Ray installation")

	installed, version, err = CheckV2Ray()
	if err != nil {
		logger.Error("Failed to verify V2Ray installation", zap.Error(err))
		return status, err
	}

	status.Installed = installed
	status.Version = version
	status.Running = IsV2RayRunning()

	logger.Info("V2Ray deployment completed",
		zap.Bool("installed", installed),
		zap.String("version", version),
		zap.Bool("running", status.Running))

	return status, nil
}

// configureV2Ray 配置V2Ray
func configureV2Ray(port int, uuid string, accessLogPath string) error {
	// V2Ray配置文件路径
	configPath := "/usr/local/etc/v2ray/config.json"
	logger.Info("Configuring V2Ray",
		zap.String("config_path", configPath),
		zap.Int("port", port))

	// 读取现有配置
	existingConfig, err := os.ReadFile(configPath)
	if err != nil {
		// 如果配置文件不存在，创建新的
		if os.IsNotExist(err) {
			logger.Info("V2Ray config file not found, creating new one", zap.String("path", configPath))
			return createNewConfig(configPath, port, uuid, accessLogPath)
		}
		logger.Error("Failed to read V2Ray config file",
			zap.String("path", configPath),
			zap.Error(err))
		return fmt.Errorf("failed to read v2ray config: %w", err)
	}

	// 检查配置是否已包含我们需要的配置
	if strings.Contains(string(existingConfig), uuid) && strings.Contains(string(existingConfig), fmt.Sprintf("%d", port)) {
		logger.Info("V2Ray config already contains required settings, skipping")
		return nil // 配置已存在，无需修改
	}

	// 创建新配置
	logger.Info("Existing V2Ray config does not match required settings, creating new config")
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
      "protocol": "vmess",
      "settings": {
        "clients": [
          {
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
	logger.Debug("Ensuring config directory exists", zap.String("dir", configDir))
	if err := os.MkdirAll(configDir, 0755); err != nil {
		logger.Error("Failed to create config directory",
			zap.String("dir", configDir),
			zap.Error(err))
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	logger.Debug("Config directory ensured", zap.String("dir", configDir))

	// 写入配置文件
	logger.Debug("Writing V2Ray config file",
		zap.String("path", configPath),
		zap.Int("config_size", len(configTemplate)))
	if err := os.WriteFile(configPath, []byte(configTemplate), 0644); err != nil {
		logger.Error("Failed to write V2Ray config file",
			zap.String("path", configPath),
			zap.Error(err))
		return fmt.Errorf("failed to write v2ray config: %w", err)
	}
	logger.Info("V2Ray config file written successfully", zap.String("path", configPath))

	// 确保日志目录存在
	logDir := "/var/log/v2ray"
	logger.Debug("Ensuring log directory exists", zap.String("dir", logDir))
	if err := os.MkdirAll(logDir, 0755); err != nil {
		logger.Error("Failed to create log directory",
			zap.String("dir", logDir),
			zap.Error(err))
		return fmt.Errorf("failed to create log directory: %w", err)
	}
	logger.Debug("Log directory ensured", zap.String("dir", logDir))

	return nil
}
