package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/yuhai94/anywhere_agent/internal/agent"
	"github.com/yuhai94/anywhere_agent/internal/config"
	"github.com/yuhai94/anywhere_agent/internal/logger"
)

func main() {
	// 解析命令行参数
	if err := config.InitCLI(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 检查是否显示版本
	if config.CLIConfig.Version {
		fmt.Printf("Anywhere Agent %s\n", config.GetVersion())
		os.Exit(0)
	}

	// 加载配置文件
	if err := config.LoadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志系统
	logger.Init(
		config.CLIConfig.LogDir,
		config.AppConfig.Log.Level,
		config.AppConfig.Log.MaxSize,
		config.AppConfig.Log.MaxBackups,
		config.AppConfig.Log.MaxAge,
	)
	defer logger.Sync()

	// 输出启动信息
	logger.Info("Starting Anywhere Agent", logger.String("version", config.GetVersion()))

	// 创建Agent实例
	logger.Info("Creating agent instance")
	agentInstance, err := agent.NewAgent(&config.AppConfig)
	if err != nil {
		logger.Fatal("Failed to create agent", zap.Error(err))
	}
	logger.Info("Agent instance created successfully")

	// 启动Agent
	logger.Info("Starting agent")
	go func() {
		if err := agentInstance.Start(); err != nil {
			logger.Fatal("Agent error", zap.Error(err))
		}
	}()

	// 等待终止信号
	logger.Info("Waiting for termination signal")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭Agent
	logger.Info("Received termination signal, stopping agent")
	agentInstance.Stop()
	logger.Info("Agent exited gracefully")
}
