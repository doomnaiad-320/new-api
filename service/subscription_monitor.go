package service

import (
	"fmt"
	"one-api/common"
	"one-api/model"
	"time"
)

// SubscriptionMonitorService 订阅监控服务
type SubscriptionMonitorService struct {
	subscriptionService *SubscriptionService
}

// NewSubscriptionMonitorService 创建订阅监控服务实例
func NewSubscriptionMonitorService() *SubscriptionMonitorService {
	return &SubscriptionMonitorService{
		subscriptionService: NewSubscriptionService(),
	}
}

// MonitorSubscriptionQuotas 监控订阅配额使用情况
func (s *SubscriptionMonitorService) MonitorSubscriptionQuotas() error {
	// 获取所有激活的订阅
	var activeSubscriptions []*model.UserSubscription
	now := time.Now().Unix()
	
	err := model.DB.Where("status = ? AND start_time <= ? AND end_time > ?", 
		model.SubscriptionStatusActive, now, now).
		Preload("User").Preload("SubscriptionPlan").
		Find(&activeSubscriptions).Error
	
	if err != nil {
		return fmt.Errorf("获取激活订阅失败: %v", err)
	}
	
	for _, subscription := range activeSubscriptions {
		err := s.checkSubscriptionQuotaWarning(subscription)
		if err != nil {
			common.SysError(fmt.Sprintf("检查订阅配额预警失败: %v", err))
		}
	}
	
	common.SysLog(fmt.Sprintf("订阅配额监控完成，检查了 %d 个激活订阅", len(activeSubscriptions)))
	return nil
}

// checkSubscriptionQuotaWarning 检查订阅配额预警
func (s *SubscriptionMonitorService) checkSubscriptionQuotaWarning(subscription *model.UserSubscription) error {
	// 检查并更新过期状态
	err := subscription.CheckAndUpdateExpiredStatus()
	if err != nil {
		return err
	}
	
	if !subscription.IsActive() {
		return nil
	}
	
	// 获取套餐配额
	planQuotas, err := subscription.SubscriptionPlan.GetModelQuotasMap()
	if err != nil {
		return err
	}
	
	// 检查每个模型的配额使用情况
	for modelName, totalQuota := range planQuotas {
		quotaInfo, err := subscription.GetModelQuotaInfo(modelName)
		if err != nil {
			continue
		}
		
		if quotaInfo.Total == 0 {
			continue
		}
		
		// 计算使用百分比
		usagePercentage := float64(quotaInfo.Used) / float64(quotaInfo.Total) * 100
		
		// 发送预警
		if usagePercentage >= 90 {
			s.sendQuotaWarning(subscription, modelName, quotaInfo, "严重预警", usagePercentage)
		} else if usagePercentage >= 80 {
			s.sendQuotaWarning(subscription, modelName, quotaInfo, "高使用率预警", usagePercentage)
		} else if usagePercentage >= 70 {
			s.sendQuotaWarning(subscription, modelName, quotaInfo, "使用率提醒", usagePercentage)
		}
	}
	
	return nil
}

// sendQuotaWarning 发送配额预警
func (s *SubscriptionMonitorService) sendQuotaWarning(subscription *model.UserSubscription, modelName string, quotaInfo *model.ModelQuotaInfo, warningType string, percentage float64) {
	message := fmt.Sprintf("%s：订阅套餐 %s 中模型 %s 的配额使用率已达 %.1f%%，剩余 %d 次，总计 %d 次", 
		warningType, subscription.SubscriptionPlan.Name, modelName, percentage, quotaInfo.Remaining, quotaInfo.Total)
	
	// 记录预警日志
	model.RecordLog(subscription.UserId, model.LogTypeSystem, message)
	
	// 这里可以扩展发送邮件、短信等通知方式
	common.SysLog(fmt.Sprintf("用户 %d 配额预警: %s", subscription.UserId, message))
}

// CleanupExpiredSubscriptions 清理过期订阅
func (s *SubscriptionMonitorService) CleanupExpiredSubscriptions() error {
	return s.subscriptionService.CleanupExpiredSubscriptions()
}

// GenerateSubscriptionReport 生成订阅报表
func (s *SubscriptionMonitorService) GenerateSubscriptionReport(startTime, endTime int64) (*SubscriptionReport, error) {
	report := &SubscriptionReport{
		StartTime: startTime,
		EndTime:   endTime,
		Stats:     make(map[string]*SubscriptionStats),
	}
	
	// 获取时间范围内的订阅数据
	var subscriptions []*model.UserSubscription
	query := model.DB.Preload("User").Preload("SubscriptionPlan")
	
	if startTime > 0 {
		query = query.Where("created_time >= ?", startTime)
	}
	if endTime > 0 {
		query = query.Where("created_time <= ?", endTime)
	}
	
	err := query.Find(&subscriptions).Error
	if err != nil {
		return nil, fmt.Errorf("获取订阅数据失败: %v", err)
	}
	
	// 统计数据
	for _, subscription := range subscriptions {
		planName := subscription.SubscriptionPlan.Name
		
		if _, exists := report.Stats[planName]; !exists {
			report.Stats[planName] = &SubscriptionStats{
				PlanName:      planName,
				TotalSales:    0,
				TotalRevenue:  0,
				ActiveCount:   0,
				ExpiredCount:  0,
				CanceledCount: 0,
			}
		}
		
		stats := report.Stats[planName]
		stats.TotalSales++
		stats.TotalRevenue += subscription.PurchasePrice
		
		switch subscription.Status {
		case model.SubscriptionStatusActive:
			stats.ActiveCount++
		case model.SubscriptionStatusExpired:
			stats.ExpiredCount++
		case model.SubscriptionStatusCanceled:
			stats.CanceledCount++
		}
	}
	
	// 计算总计
	for _, stats := range report.Stats {
		report.TotalSales += stats.TotalSales
		report.TotalRevenue += stats.TotalRevenue
		report.TotalActive += stats.ActiveCount
		report.TotalExpired += stats.ExpiredCount
		report.TotalCanceled += stats.CanceledCount
	}
	
	return report, nil
}

// GetUserSubscriptionSummary 获取用户订阅摘要
func (s *SubscriptionMonitorService) GetUserSubscriptionSummary(userId int) (*UserSubscriptionSummary, error) {
	summary := &UserSubscriptionSummary{
		UserId:        userId,
		TotalQuotas:   make(map[string]*model.ModelQuotaInfo),
		Subscriptions: make([]*SubscriptionInfo, 0),
	}
	
	// 获取用户的激活订阅
	subscriptions, err := model.GetActiveUserSubscriptions(userId)
	if err != nil {
		return nil, err
	}
	
	summary.ActiveCount = len(subscriptions)
	
	var nearestExpiry int64 = 0
	
	for _, subscription := range subscriptions {
		// 检查并更新过期状态
		err := subscription.CheckAndUpdateExpiredStatus()
		if err != nil {
			continue
		}
		
		if !subscription.IsActive() {
			continue
		}
		
		// 找到最近的过期时间
		if nearestExpiry == 0 || subscription.EndTime < nearestExpiry {
			nearestExpiry = subscription.EndTime
		}
		
		// 汇总配额
		quotas, err := subscription.GetModelQuotasMap()
		if err != nil {
			continue
		}
		
		subscriptionInfo := &SubscriptionInfo{
			Id:           subscription.Id,
			PlanName:     subscription.SubscriptionPlan.Name,
			StartTime:    subscription.StartTime,
			EndTime:      subscription.EndTime,
			Status:       subscription.GetStatusText(),
			ModelQuotas:  make(map[string]*model.ModelQuotaInfo),
		}
		
		for modelName := range quotas {
			quotaInfo, err := subscription.GetModelQuotaInfo(modelName)
			if err != nil {
				continue
			}
			
			subscriptionInfo.ModelQuotas[modelName] = quotaInfo
			
			if existing, exists := summary.TotalQuotas[modelName]; exists {
				existing.Total += quotaInfo.Total
				existing.Used += quotaInfo.Used
				existing.Remaining += quotaInfo.Remaining
			} else {
				summary.TotalQuotas[modelName] = &model.ModelQuotaInfo{
					Total:     quotaInfo.Total,
					Used:      quotaInfo.Used,
					Remaining: quotaInfo.Remaining,
				}
			}
		}
		
		summary.Subscriptions = append(summary.Subscriptions, subscriptionInfo)
	}
	
	summary.NearestExpiry = nearestExpiry
	
	return summary, nil
}

// GetSystemSubscriptionStats 获取系统订阅统计
func (s *SubscriptionMonitorService) GetSystemSubscriptionStats() (*SystemSubscriptionStats, error) {
	stats := &SystemSubscriptionStats{
		PlanStats: make(map[string]*PlanStats),
	}
	
	// 统计各状态的订阅数量
	err := model.DB.Model(&model.UserSubscription{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&stats.StatusCounts).Error
	if err != nil {
		return nil, err
	}
	
	// 统计各套餐的订阅情况
	var planResults []struct {
		PlanId    int     `json:"plan_id"`
		PlanName  string  `json:"plan_name"`
		Count     int     `json:"count"`
		Revenue   float64 `json:"revenue"`
	}
	
	err = model.DB.Table("user_subscriptions").
		Select("subscription_plan_id as plan_id, subscription_plans.name as plan_name, COUNT(*) as count, SUM(purchase_price) as revenue").
		Joins("LEFT JOIN subscription_plans ON user_subscriptions.subscription_plan_id = subscription_plans.id").
		Group("subscription_plan_id, subscription_plans.name").
		Scan(&planResults).Error
	if err != nil {
		return nil, err
	}
	
	for _, result := range planResults {
		stats.PlanStats[result.PlanName] = &PlanStats{
			PlanId:      result.PlanId,
			PlanName:    result.PlanName,
			TotalSales:  result.Count,
			TotalRevenue: result.Revenue,
		}
		
		stats.TotalSubscriptions += result.Count
		stats.TotalRevenue += result.Revenue
	}
	
	return stats, nil
}

// SubscriptionReport 订阅报表
type SubscriptionReport struct {
	StartTime      int64                        `json:"start_time"`
	EndTime        int64                        `json:"end_time"`
	TotalSales     int                          `json:"total_sales"`
	TotalRevenue   float64                      `json:"total_revenue"`
	TotalActive    int                          `json:"total_active"`
	TotalExpired   int                          `json:"total_expired"`
	TotalCanceled  int                          `json:"total_canceled"`
	Stats          map[string]*SubscriptionStats `json:"stats"`
}

// SubscriptionStats 订阅统计
type SubscriptionStats struct {
	PlanName      string  `json:"plan_name"`
	TotalSales    int     `json:"total_sales"`
	TotalRevenue  float64 `json:"total_revenue"`
	ActiveCount   int     `json:"active_count"`
	ExpiredCount  int     `json:"expired_count"`
	CanceledCount int     `json:"canceled_count"`
}

// UserSubscriptionSummary 用户订阅摘要
type UserSubscriptionSummary struct {
	UserId        int                              `json:"user_id"`
	ActiveCount   int                              `json:"active_count"`
	TotalQuotas   map[string]*model.ModelQuotaInfo `json:"total_quotas"`
	NearestExpiry int64                            `json:"nearest_expiry"`
	Subscriptions []*SubscriptionInfo              `json:"subscriptions"`
}

// SubscriptionInfo 订阅信息
type SubscriptionInfo struct {
	Id          int                              `json:"id"`
	PlanName    string                           `json:"plan_name"`
	StartTime   int64                            `json:"start_time"`
	EndTime     int64                            `json:"end_time"`
	Status      string                           `json:"status"`
	ModelQuotas map[string]*model.ModelQuotaInfo `json:"model_quotas"`
}

// SystemSubscriptionStats 系统订阅统计
type SystemSubscriptionStats struct {
	TotalSubscriptions int                    `json:"total_subscriptions"`
	TotalRevenue       float64               `json:"total_revenue"`
	StatusCounts       []StatusCount         `json:"status_counts"`
	PlanStats          map[string]*PlanStats `json:"plan_stats"`
}

// StatusCount 状态统计
type StatusCount struct {
	Status int `json:"status"`
	Count  int `json:"count"`
}

// PlanStats 套餐统计
type PlanStats struct {
	PlanId       int     `json:"plan_id"`
	PlanName     string  `json:"plan_name"`
	TotalSales   int     `json:"total_sales"`
	TotalRevenue float64 `json:"total_revenue"`
}
