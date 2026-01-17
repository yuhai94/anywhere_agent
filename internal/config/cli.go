package config

import (
	"flag"
	"fmt"
	"os"
)

// CLIConfig 存储命令行参数
var CLIConfig struct {
	ConfigFile string
	LogDir     string
	Version    bool
}

// GetVersion 返回版本信息
func GetVersion() string {
	return "v1.0.0"
}

// InitCLI 初始化命令行参数
func InitCLI() error {
	// 使用标准库flag解析命令行参数
	flag.StringVar(&CLIConfig.ConfigFile, "config", "./config.yaml", "Config file path")
	flag.StringVar(&CLIConfig.LogDir, "log-dir", "/var/log/aw_agent/", "Log directory")
	flag.BoolVar(&CLIConfig.Version, "version", false, "Show version information")

	// 自定义help信息
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Anywhere Agent is a tool for deploying and managing V2Ray on EC2 instances.\n")
		fmt.Fprintf(os.Stderr, "It automatically checks and deploys V2Ray, monitors traffic, and manages EC2 instances.\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	// 解析命令行参数
	flag.Parse()

	return nil
}
