package model

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// SubscriptionUsage 订阅使用记录表
type SubscriptionUsage struct {
	Id                 int    `json:"id" gorm:"primaryKey"`
	UserId             int    `json:"user_id" gorm:"index;not null"`                              // 用户ID
	UserSubscriptionId int    `json:"user_subscription_id" gorm:"index;not null"`                // 用户订阅ID
	ModelName          string `json:"model_name" gorm:"type:varchar(100);index;not null"`        // 模型名称
	UsageCount         int    `json:"usage_count" gorm:"default:1"`                              // 使用次数
	TokensUsed         int    `json:"tokens_used" gorm:"default:0"`                              // 使用的token数量
	RequestId          string `json:"request_id" gorm:"type:varchar(100);index"`                 // 请求ID
	CreatedTime        int64  `json:"created_time" gorm:"bigint;index;autoCreateTime"`           // 创建时间
	
	// 关联字段
	User             *User             `json:"user,omitempty" gorm:"foreignKey:UserId"`
	UserSubscription *UserSubscription `json:"user_subscription,omitempty" gorm:"foreignKey:UserSubscriptionId"`
}

// Insert 创建使用记录
func (su *SubscriptionUsage) Insert() error {
	if su.UserId == 0 {
		return errors.New("用户ID不能为空")
	}
	if su.UserSubscriptionId == 0 {
		return errors.New("用户订阅ID不能为空")
	}
	if su.ModelName == "" {
		return errors.New("模型名称不能为空")
	}
	
	su.CreatedTime = time.Now().Unix()
	return DB.Create(su).Error
}

// GetUsageBySubscription 获取订阅的使用记录
func GetUsageBySubscription(subscriptionId int, page, pageSize int) ([]*SubscriptionUsage, int64, error) {
	var usages []*SubscriptionUsage
	var total int64
	
	query := DB.Model(&SubscriptionUsage{}).Where("user_subscription_id = ?", subscriptionId)
	
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * pageSize
	err = query.Preload("User").
		Order("created_time DESC").
		Offset(offset).Limit(pageSize).
		Find(&usages).Error
	
	return usages, total, err
}

// GetUsageByUser 获取用户的使用记录
func GetUsageByUser(userId int, page, pageSize int) ([]*SubscriptionUsage, int64, error) {
	var usages []*SubscriptionUsage
	var total int64
	
	query := DB.Model(&SubscriptionUsage{}).Where("user_id = ?", userId)
	
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * pageSize
	err = query.Preload("UserSubscription").Preload("UserSubscription.SubscriptionPlan").
		Order("created_time DESC").
		Offset(offset).Limit(pageSize).
		Find(&usages).Error
	
	return usages, total, err
}

// GetUsageStatsBySubscription 获取订阅的使用统计
func GetUsageStatsBySubscription(subscriptionId int) (map[string]*UsageStats, error) {
	var results []struct {
		ModelName  string `json:"model_name"`
		TotalCount int    `json:"total_count"`
		TotalTokens int   `json:"total_tokens"`
	}
	
	err := DB.Model(&SubscriptionUsage{}).
		Select("model_name, SUM(usage_count) as total_count, SUM(tokens_used) as total_tokens").
		Where("user_subscription_id = ?", subscriptionId).
		Group("model_name").
		Find(&results).Error
	
	if err != nil {
		return nil, err
	}
	
	stats := make(map[string]*UsageStats)
	for _, result := range results {
		stats[result.ModelName] = &UsageStats{
			ModelName:   result.ModelName,
			TotalCount:  result.TotalCount,
			TotalTokens: result.TotalTokens,
		}
	}
	
	return stats, nil
}

// GetUsageStatsByUser 获取用户的使用统计
func GetUsageStatsByUser(userId int, startTime, endTime int64) (map[string]*UsageStats, error) {
	var results []struct {
		ModelName  string `json:"model_name"`
		TotalCount int    `json:"total_count"`
		TotalTokens int   `json:"total_tokens"`
	}
	
	query := DB.Model(&SubscriptionUsage{}).
		Select("model_name, SUM(usage_count) as total_count, SUM(tokens_used) as total_tokens").
		Where("user_id = ?", userId)
	
	if startTime > 0 {
		query = query.Where("created_time >= ?", startTime)
	}
	if endTime > 0 {
		query = query.Where("created_time <= ?", endTime)
	}
	
	err := query.Group("model_name").Find(&results).Error
	if err != nil {
		return nil, err
	}
	
	stats := make(map[string]*UsageStats)
	for _, result := range results {
		stats[result.ModelName] = &UsageStats{
			ModelName:   result.ModelName,
			TotalCount:  result.TotalCount,
			TotalTokens: result.TotalTokens,
		}
	}
	
	return stats, nil
}

// UsageStats 使用统计
type UsageStats struct {
	ModelName   string `json:"model_name"`
	TotalCount  int    `json:"total_count"`
	TotalTokens int    `json:"total_tokens"`
}

// GetDailyUsageStats 获取每日使用统计
func GetDailyUsageStats(userId int, days int) ([]*DailyUsageStats, error) {
	var results []*DailyUsageStats
	
	// 计算开始时间（days天前的0点）
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day()-days+1, 0, 0, 0, 0, now.Location()).Unix()
	
	query := `
		SELECT 
			DATE(FROM_UNIXTIME(created_time)) as date,
			model_name,
			SUM(usage_count) as total_count,
			SUM(tokens_used) as total_tokens
		FROM subscription_usages 
		WHERE user_id = ? AND created_time >= ?
		GROUP BY DATE(FROM_UNIXTIME(created_time)), model_name
		ORDER BY date DESC, model_name
	`
	
	err := DB.Raw(query, userId, startTime).Scan(&results).Error
	return results, err
}

// DailyUsageStats 每日使用统计
type DailyUsageStats struct {
	Date        string `json:"date"`
	ModelName   string `json:"model_name"`
	TotalCount  int    `json:"total_count"`
	TotalTokens int    `json:"total_tokens"`
}

// RecordSubscriptionUsage 记录订阅使用
func RecordSubscriptionUsage(userId, subscriptionId int, modelName string, usageCount, tokensUsed int, requestId string) error {
	usage := &SubscriptionUsage{
		UserId:             userId,
		UserSubscriptionId: subscriptionId,
		ModelName:          modelName,
		UsageCount:         usageCount,
		TokensUsed:         tokensUsed,
		RequestId:          requestId,
	}
	
	return usage.Insert()
}

// GetModelUsageToday 获取今日模型使用量
func GetModelUsageToday(userId int, modelName string) (int, error) {
	// 获取今日0点时间戳
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	
	var totalCount int64
	err := DB.Model(&SubscriptionUsage{}).
		Where("user_id = ? AND model_name = ? AND created_time >= ?", userId, modelName, todayStart).
		Select("COALESCE(SUM(usage_count), 0)").
		Scan(&totalCount).Error
	
	return int(totalCount), err
}

// GetModelUsageThisMonth 获取本月模型使用量
func GetModelUsageThisMonth(userId int, modelName string) (int, error) {
	// 获取本月1号0点时间戳
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()
	
	var totalCount int64
	err := DB.Model(&SubscriptionUsage{}).
		Where("user_id = ? AND model_name = ? AND created_time >= ?", userId, modelName, monthStart).
		Select("COALESCE(SUM(usage_count), 0)").
		Scan(&totalCount).Error
	
	return int(totalCount), err
}

// CleanupOldUsageRecords 清理旧的使用记录（保留指定天数）
func CleanupOldUsageRecords(retentionDays int) error {
	if retentionDays <= 0 {
		return errors.New("保留天数必须大于0")
	}
	
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays).Unix()
	
	result := DB.Where("created_time < ?", cutoffTime).Delete(&SubscriptionUsage{})
	if result.Error != nil {
		return result.Error
	}
	
	fmt.Printf("清理了 %d 条旧的使用记录\n", result.RowsAffected)
	return nil
}

// GetTopModelsUsage 获取使用量最高的模型
func GetTopModelsUsage(userId int, limit int, startTime, endTime int64) ([]*UsageStats, error) {
	var results []*UsageStats
	
	query := DB.Model(&SubscriptionUsage{}).
		Select("model_name, SUM(usage_count) as total_count, SUM(tokens_used) as total_tokens").
		Where("user_id = ?", userId)
	
	if startTime > 0 {
		query = query.Where("created_time >= ?", startTime)
	}
	if endTime > 0 {
		query = query.Where("created_time <= ?", endTime)
	}
	
	err := query.Group("model_name").
		Order("total_count DESC").
		Limit(limit).
		Find(&results).Error
	
	return results, err
}

// GetUsageByDateRange 获取指定时间范围的使用记录
func GetUsageByDateRange(userId int, startTime, endTime int64, page, pageSize int) ([]*SubscriptionUsage, int64, error) {
	var usages []*SubscriptionUsage
	var total int64
	
	query := DB.Model(&SubscriptionUsage{}).Where("user_id = ?", userId)
	
	if startTime > 0 {
		query = query.Where("created_time >= ?", startTime)
	}
	if endTime > 0 {
		query = query.Where("created_time <= ?", endTime)
	}
	
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * pageSize
	err = query.Preload("UserSubscription").Preload("UserSubscription.SubscriptionPlan").
		Order("created_time DESC").
		Offset(offset).Limit(pageSize).
		Find(&usages).Error
	
	return usages, total, err
}
