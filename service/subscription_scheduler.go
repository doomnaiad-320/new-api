package service

import (
	"one-api/common"
	"one-api/model"
	"time"
)

// SubscriptionScheduler 订阅调度器
type SubscriptionScheduler struct {
	monitorService *SubscriptionMonitorService
	ticker         *time.Ticker
	stopChan       chan bool
}

// NewSubscriptionScheduler 创建订阅调度器
func NewSubscriptionScheduler() *SubscriptionScheduler {
	return &SubscriptionScheduler{
		monitorService: NewSubscriptionMonitorService(),
		stopChan:       make(chan bool),
	}
}

// Start 启动订阅调度器
func (s *SubscriptionScheduler) Start() {
	// 每小时执行一次监控任务
	s.ticker = time.NewTicker(time.Hour)
	
	common.SysLog("订阅调度器已启动，每小时执行一次监控任务")
	
	go func() {
		// 立即执行一次
		s.runMonitorTasks()
		
		for {
			select {
			case <-s.ticker.C:
				s.runMonitorTasks()
			case <-s.stopChan:
				common.SysLog("订阅调度器已停止")
				return
			}
		}
	}()
}

// Stop 停止订阅调度器
func (s *SubscriptionScheduler) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.stopChan <- true
}

// runMonitorTasks 执行监控任务
func (s *SubscriptionScheduler) runMonitorTasks() {
	common.SysLog("开始执行订阅监控任务")
	
	// 1. 监控订阅配额使用情况
	err := s.monitorService.MonitorSubscriptionQuotas()
	if err != nil {
		common.SysError("监控订阅配额失败: " + err.Error())
	}
	
	// 2. 清理过期订阅
	err = s.monitorService.CleanupExpiredSubscriptions()
	if err != nil {
		common.SysError("清理过期订阅失败: " + err.Error())
	}
	
	common.SysLog("订阅监控任务执行完成")
}

// RunDailyCleanup 执行每日清理任务
func (s *SubscriptionScheduler) RunDailyCleanup() {
	common.SysLog("开始执行每日清理任务")
	
	// 清理旧的使用记录（保留90天）
	err := model.CleanupOldUsageRecords(90)
	if err != nil {
		common.SysError("清理旧使用记录失败: " + err.Error())
	}
	
	common.SysLog("每日清理任务执行完成")
}

// StartDailyScheduler 启动每日调度器
func (s *SubscriptionScheduler) StartDailyScheduler() {
	// 每天凌晨2点执行清理任务
	go func() {
		for {
			now := time.Now()
			// 计算到明天凌晨2点的时间
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 2, 0, 0, 0, now.Location())
			duration := next.Sub(now)
			
			common.SysLog("每日清理任务将在 " + duration.String() + " 后执行")
			
			timer := time.NewTimer(duration)
			select {
			case <-timer.C:
				s.RunDailyCleanup()
			case <-s.stopChan:
				timer.Stop()
				return
			}
		}
	}()
}
