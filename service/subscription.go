package service

import (
	"errors"
	"fmt"
	"one-api/common"
	"one-api/model"
	relaycommon "one-api/relay/common"
	"time"

	"github.com/gin-gonic/gin"
)

// SubscriptionService 订阅服务
type SubscriptionService struct{}

// NewSubscriptionService 创建订阅服务实例
func NewSubscriptionService() *SubscriptionService {
	return &SubscriptionService{}
}

// CheckAndConsumeSubscriptionQuota 检查并消费订阅配额
// 返回值：是否使用了订阅配额，消费的配额数量，错误信息
func (s *SubscriptionService) CheckAndConsumeSubscriptionQuota(ctx *gin.Context, relayInfo *relaycommon.RelayInfo, modelName string, usageCount int) (bool, int, error) {
	userId := relayInfo.UserId
	if userId == 0 {
		return false, 0, errors.New("用户ID为空")
	}

	// 获取用户的激活订阅
	subscriptions, err := model.GetActiveUserSubscriptions(userId)
	if err != nil {
		common.SysError(fmt.Sprintf("获取用户订阅失败: %v", err))
		return false, 0, nil // 不阻断请求，继续使用原有计费方式
	}

	if len(subscriptions) == 0 {
		// 没有激活的订阅，使用原有计费方式
		return false, 0, nil
	}

	// 查找包含该模型的订阅，按结束时间排序（优先使用即将过期的）
	var availableSubscription *model.UserSubscription
	var maxQuota int

	for _, subscription := range subscriptions {
		// 检查并更新过期状态
		err := subscription.CheckAndUpdateExpiredStatus()
		if err != nil {
			common.SysError(fmt.Sprintf("更新订阅状态失败: %v", err))
			continue
		}

		if !subscription.IsActive() {
			continue
		}

		// 获取该模型的配额信息
		quotaInfo, err := subscription.GetModelQuotaInfo(modelName)
		if err != nil {
			common.SysError(fmt.Sprintf("获取模型配额信息失败: %v", err))
			continue
		}

		if quotaInfo.Remaining >= usageCount {
			if availableSubscription == nil || quotaInfo.Remaining > maxQuota {
				availableSubscription = subscription
				maxQuota = quotaInfo.Remaining
			}
		}
	}

	if availableSubscription == nil {
		// 没有足够配额的订阅，使用原有计费方式
		return false, 0, nil
	}

	// 消费订阅配额
	err = availableSubscription.ConsumeModelQuota(modelName, usageCount)
	if err != nil {
		common.SysError(fmt.Sprintf("消费订阅配额失败: %v", err))
		return false, 0, nil // 不阻断请求，继续使用原有计费方式
	}

	// 记录使用记录
	err = model.RecordSubscriptionUsage(userId, availableSubscription.Id, modelName, usageCount, 0, relayInfo.RequestId)
	if err != nil {
		common.SysError(fmt.Sprintf("记录订阅使用失败: %v", err))
		// 这里不回滚配额消费，因为记录失败不应该影响用户使用
	}

	// 记录日志
	model.RecordLog(userId, model.LogTypeConsumeQuota,
		fmt.Sprintf("使用订阅配额: 模型 %s，次数 %d，订阅ID %d", modelName, usageCount, availableSubscription.Id))

	// 设置RelayInfo标记（只在实际消费时设置）
	if relayInfo != nil {
		relayInfo.SubscriptionId = availableSubscription.Id
	}

	return true, usageCount, nil
}

// GetUserSubscriptionQuotas 获取用户所有订阅的配额信息
func (s *SubscriptionService) GetUserSubscriptionQuotas(userId int) (map[string]*model.ModelQuotaInfo, error) {
	subscriptions, err := model.GetActiveUserSubscriptions(userId)
	if err != nil {
		return nil, err
	}

	totalQuotas := make(map[string]*model.ModelQuotaInfo)

	for _, subscription := range subscriptions {
		// 检查并更新过期状态
		err := subscription.CheckAndUpdateExpiredStatus()
		if err != nil {
			common.SysError(fmt.Sprintf("更新订阅状态失败: %v", err))
			continue
		}

		if !subscription.IsActive() {
			continue
		}

		quotas, err := subscription.GetModelQuotasMap()
		if err != nil {
			continue
		}

		for modelName := range quotas {
			quotaInfo, err := subscription.GetModelQuotaInfo(modelName)
			if err != nil {
				continue
			}

			if existing, exists := totalQuotas[modelName]; exists {
				existing.Total += quotaInfo.Total
				existing.Used += quotaInfo.Used
				existing.Remaining += quotaInfo.Remaining
			} else {
				totalQuotas[modelName] = quotaInfo
			}
		}
	}

	return totalQuotas, nil
}

// CheckModelQuotaAvailable 检查模型配额是否可用
func (s *SubscriptionService) CheckModelQuotaAvailable(userId int, modelName string, requiredCount int) (bool, *model.ModelQuotaInfo, error) {
	subscriptions, err := model.GetActiveUserSubscriptions(userId)
	if err != nil {
		return false, nil, err
	}

	totalRemaining := 0
	totalUsed := 0
	totalQuota := 0

	for _, subscription := range subscriptions {
		// 检查并更新过期状态
		err := subscription.CheckAndUpdateExpiredStatus()
		if err != nil {
			continue
		}

		if !subscription.IsActive() {
			continue
		}

		quotaInfo, err := subscription.GetModelQuotaInfo(modelName)
		if err != nil {
			continue
		}

		totalQuota += quotaInfo.Total
		totalUsed += quotaInfo.Used
		totalRemaining += quotaInfo.Remaining
	}

	quotaInfo := &model.ModelQuotaInfo{
		Total:     totalQuota,
		Used:      totalUsed,
		Remaining: totalRemaining,
	}

	return totalRemaining >= requiredCount, quotaInfo, nil
}

// GetSubscriptionUsageStats 获取订阅使用统计
func (s *SubscriptionService) GetSubscriptionUsageStats(userId int, days int) (map[string]*model.UsageStats, error) {
	endTime := time.Now().Unix()
	startTime := endTime - int64(days*24*3600)

	return model.GetUsageStatsByUser(userId, startTime, endTime)
}

// CleanupExpiredSubscriptions 清理过期订阅
func (s *SubscriptionService) CleanupExpiredSubscriptions() error {
	// 这个方法可以定期调用来清理过期订阅
	// 获取所有激活但实际已过期的订阅
	var expiredSubscriptions []*model.UserSubscription
	now := time.Now().Unix()

	err := model.DB.Where("status = ? AND end_time <= ?", model.SubscriptionStatusActive, now).
		Find(&expiredSubscriptions).Error
	if err != nil {
		return err
	}

	for _, subscription := range expiredSubscriptions {
		subscription.Status = model.SubscriptionStatusExpired
		err := subscription.Update()
		if err != nil {
			common.SysError(fmt.Sprintf("更新过期订阅状态失败: %v", err))
		}
	}

	common.SysLog(fmt.Sprintf("清理了 %d 个过期订阅", len(expiredSubscriptions)))
	return nil
}

// SendQuotaWarning 发送配额预警
func (s *SubscriptionService) SendQuotaWarning(userId int, modelName string, remaining int, total int) {
	if total == 0 {
		return
	}

	percentage := float64(remaining) / float64(total) * 100

	var warningMessage string
	if percentage <= 10 {
		warningMessage = fmt.Sprintf("⚠️ 配额预警：模型 %s 的订阅配额即将用完，剩余 %d 次（%.1f%%）", modelName, remaining, percentage)
	} else if percentage <= 20 {
		warningMessage = fmt.Sprintf("📊 配额提醒：模型 %s 的订阅配额剩余 %d 次（%.1f%%）", modelName, remaining, percentage)
	}

	if warningMessage != "" {
		// 记录预警日志
		model.RecordLog(userId, model.LogTypeSystem, warningMessage)
		
		// 这里可以扩展发送邮件或其他通知方式
		common.SysLog(fmt.Sprintf("用户 %d 配额预警: %s", userId, warningMessage))
	}
}

// ValidateSubscriptionPlan 验证订阅套餐
func (s *SubscriptionService) ValidateSubscriptionPlan(plan *model.SubscriptionPlan) error {
	if plan.Name == "" {
		return errors.New("套餐名称不能为空")
	}

	if plan.Price < 0 {
		return errors.New("套餐价格不能为负数")
	}

	if plan.Duration <= 0 {
		return errors.New("有效期必须大于0天")
	}

	// 验证模型配额
	quotas, err := plan.GetModelQuotasMap()
	if err != nil {
		return fmt.Errorf("模型配额格式错误: %v", err)
	}

	if len(quotas) == 0 {
		return errors.New("至少需要包含一个模型配额")
	}

	for modelName, quota := range quotas {
		if modelName == "" {
			return errors.New("模型名称不能为空")
		}
		if quota < 0 {
			return fmt.Errorf("模型 %s 的配额不能为负数", modelName)
		}
	}

	return nil
}

// GetSubscriptionSummary 获取订阅摘要信息
func (s *SubscriptionService) GetSubscriptionSummary(userId int) (*SubscriptionSummary, error) {
	subscriptions, err := model.GetActiveUserSubscriptions(userId)
	if err != nil {
		return nil, err
	}

	summary := &SubscriptionSummary{
		ActiveCount: len(subscriptions),
		TotalQuotas: make(map[string]*model.ModelQuotaInfo),
	}

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

		for modelName := range quotas {
			quotaInfo, err := subscription.GetModelQuotaInfo(modelName)
			if err != nil {
				continue
			}

			if existing, exists := summary.TotalQuotas[modelName]; exists {
				existing.Total += quotaInfo.Total
				existing.Used += quotaInfo.Used
				existing.Remaining += quotaInfo.Remaining
			} else {
				summary.TotalQuotas[modelName] = quotaInfo
			}
		}
	}

	summary.NearestExpiry = nearestExpiry

	return summary, nil
}

// SubscriptionSummary 订阅摘要
type SubscriptionSummary struct {
	ActiveCount   int                              `json:"active_count"`
	TotalQuotas   map[string]*model.ModelQuotaInfo `json:"total_quotas"`
	NearestExpiry int64                            `json:"nearest_expiry"`
}
