package agent

import (
	"time"

	"github.com/yuhai94/anywhere_agent/internal/aws"
	"github.com/yuhai94/anywhere_agent/internal/config"
	"github.com/yuhai94/anywhere_agent/internal/logger"
	"github.com/yuhai94/anywhere_agent/internal/v2ray"
	"go.uber.org/zap"
)

// Scheduler 调度器，定期执行任务
type Scheduler struct {
	config     *config.Config
	ec2Client  *aws.EC2Client
	stats      *v2ray.TrafficMonitor
	deployChan chan *v2ray.DeployStatus
	stopChan   chan struct{}
	isRunning  bool
}

// NewScheduler 创建新的调度器
func NewScheduler(cfg *config.Config, ec2Client *aws.EC2Client, stats *v2ray.TrafficMonitor, deployChan chan *v2ray.DeployStatus) *Scheduler {
	return &Scheduler{
		config:     cfg,
		ec2Client:  ec2Client,
		stats:      stats,
		deployChan: deployChan,
		stopChan:   make(chan struct{}),
		isRunning:  false,
	}
}

// Start 启动调度器
func (s *Scheduler) Start() {
	if s.isRunning {
		return
	}

	s.isRunning = true
	logger.Info("Starting scheduler...")

	// 启动实例删除检查协程
	go s.instanceDeleteLoop()
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	if !s.isRunning {
		return
	}

	logger.Info("Stopping scheduler...")
	close(s.stopChan)
	s.isRunning = false
}

// instanceDeleteLoop 实例删除检查循环
func (s *Scheduler) instanceDeleteLoop() {
	// 从配置获取实例删除检查间隔
	checkInterval := s.config.Checks.TrafficInterval
	logger.Info("Setting instance delete check interval", zap.Int("minutes", checkInterval))

	ticker := time.NewTicker(time.Duration(checkInterval) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 检查是否空闲
			idle, err := s.stats.IsIdle()
			if err != nil {
				logger.Error("Failed to check if instance is idle", zap.Error(err))
				continue
			}

			if idle {
				logger.Info("Instance is idle for 30 minutes, terminating...")

				// 获取实例ID
				instanceID, err := aws.GetInstanceID()
				if err != nil {
					logger.Error("Failed to get instance ID", zap.Error(err))
					continue
				}

				// 终止实例
				if err := s.ec2Client.TerminateInstance(instanceID); err != nil {
					logger.Error("Failed to terminate instance", zap.Error(err))
					continue
				}

				logger.Info("Instance terminated successfully", logger.String("instance_id", instanceID))
			}

		case <-s.stopChan:
			return
		}
	}
}
