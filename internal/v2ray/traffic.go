package v2ray

import (
	"os"
	"time"
)

// TrafficStats 流量统计信息
type TrafficStats struct {
	LastActive time.Time `json:"last_active"` // 最后活动时间（文件修改时间）
	HasTraffic bool      `json:"has_traffic"` // 是否有流量
}

// TrafficMonitor 流量监控器
type TrafficMonitor struct {
	logPath     string
	idleTimeout int
}

// NewTrafficMonitor 创建新的流量监控器
func NewTrafficMonitor(logPath string, idleTimeout int) *TrafficMonitor {
	return &TrafficMonitor{
		logPath:     logPath,
		idleTimeout: idleTimeout,
	}
}

// CheckTraffic 检查最近30分钟的流量
func (tm *TrafficMonitor) CheckTraffic() (*TrafficStats, error) {
	// 检查access.log文件的修改时间
	fileInfo, err := os.Stat(tm.logPath)
	var lastActive time.Time
	var hasTraffic bool

	if err != nil {
		if os.IsNotExist(err) {
			// 日志文件不存在，返回无流量
			return &TrafficStats{
				LastActive: time.Time{},
				HasTraffic: false,
			}, nil
		}
		return nil, err
	}

	// 获取文件修改时间
	lastActive = fileInfo.ModTime()
	hasTraffic = true

	return &TrafficStats{
		LastActive: lastActive,
		HasTraffic: hasTraffic,
	}, nil
}

// IsIdle 检查是否处于空闲状态
func (tm *TrafficMonitor) IsIdle() (bool, error) {
	stats, err := tm.CheckTraffic()
	if err != nil {
		return false, err
	}

	// 如果没有流量，返回空闲
	if !stats.HasTraffic {
		return true, nil
	}

	// 检查最后活动时间是否超过空闲超时
	idleTime := time.Since(stats.LastActive)
	return idleTime > time.Duration(tm.idleTimeout)*time.Second, nil
}
