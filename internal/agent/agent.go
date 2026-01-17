package agent

import (
	"sync"

	"github.com/yuhai94/anywhere_agent/internal/api"
	"github.com/yuhai94/anywhere_agent/internal/aws"
	"github.com/yuhai94/anywhere_agent/internal/config"
	"github.com/yuhai94/anywhere_agent/internal/logger"
	"github.com/yuhai94/anywhere_agent/internal/v2ray"
	"go.uber.org/zap"
)

// Agent Anywhere Agent核心结构
type Agent struct {
	config     *config.Config
	ec2Client  *aws.EC2Client
	apiServer  *api.APIServer
	scheduler  *Scheduler
	stats      *v2ray.TrafficMonitor
	deployChan chan *v2ray.DeployStatus
	wg         sync.WaitGroup
	stopChan   chan struct{}
}

// NewAgent 创建新的Agent实例
func NewAgent(cfg *config.Config) (*Agent, error) {
	// 创建部署状态通道
	deployChan := make(chan *v2ray.DeployStatus, 1)

	// 创建流量监控器
	stats := v2ray.NewTrafficMonitor(cfg.V2Ray.AccessLog, cfg.Checks.IdleTimeout)

	// 创建AWS EC2客户端
	ec2Client, err := aws.NewEC2Client()
	if err != nil {
		return nil, err
	}

	// 创建API服务器
	apiServer := api.NewAPIServer(cfg, deployChan, stats)

	// 创建调度器
	scheduler := NewScheduler(cfg, ec2Client, stats, deployChan)

	return &Agent{
		config:     cfg,
		ec2Client:  ec2Client,
		apiServer:  apiServer,
		scheduler:  scheduler,
		stats:      stats,
		deployChan: deployChan,
		stopChan:   make(chan struct{}),
	}, nil
}

// Start 启动Agent
func (a *Agent) Start() error {
	logger.Info("Starting Anywhere Agent...")

	// 1. 部署V2Ray
	go a.deployV2Ray()

	// 2. 启动API服务器
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := a.apiServer.Start(); err != nil {
			logger.Error("API server error", zap.Error(err))
		}
	}()

	// 3. 启动调度器
	a.scheduler.Start()

	logger.Info("Anywhere Agent started successfully")

	// 等待停止信号
	a.wg.Wait()

	return nil
}

// Stop 停止Agent
func (a *Agent) Stop() {
	logger.Info("Stopping Anywhere Agent...")

	// 关闭停止通道
	close(a.stopChan)

	// 停止调度器
	a.scheduler.Stop()

	// 等待所有goroutine完成
	a.wg.Wait()

	logger.Info("Anywhere Agent stopped successfully")
}

// deployV2Ray 部署V2Ray
func (a *Agent) deployV2Ray() {
	logger.Info("Deploying V2Ray...")

	// 检查V2Ray是否已安装
	installed, version, err := v2ray.CheckV2Ray()
	if err != nil {
		logger.Error("Failed to check V2Ray", zap.Error(err))
		return
	}

	if installed {
		logger.Info("V2Ray already installed", logger.String("version", version))
		// 发送部署状态
		status := &v2ray.DeployStatus{
			Installed: true,
			Running:   v2ray.IsV2RayRunning(),
			Version:   version,
			Progress:  100,
			Message:   "V2Ray already installed",
		}
		select {
		case a.deployChan <- status:
		default:
			// 通道已满，忽略
		}
		return
	}

	// 部署V2Ray，传递access log路径
	status, err := v2ray.DeployV2Ray(a.config.V2Ray.Port, a.config.V2Ray.UUID, a.config.V2Ray.AccessLog)
	if err != nil {
		logger.Error("Failed to deploy V2Ray", zap.Error(err))
		status.Message = err.Error()
	}

	// 发送部署状态
	select {
	case a.deployChan <- status:
	default:
		// 通道已满，忽略
	}

	logger.Info("V2Ray deployment completed")
}
