package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"one-api/common"
	"time"

	"gorm.io/gorm"
)

// UserSubscription 用户订阅表
type UserSubscription struct {
	Id                 int            `json:"id" gorm:"primaryKey"`
	UserId             int            `json:"user_id" gorm:"index;not null"`                              // 用户ID
	SubscriptionPlanId int            `json:"subscription_plan_id" gorm:"index;not null"`                // 订阅套餐ID
	Status             int            `json:"status" gorm:"default:1;index"`                             // 状态：1激活，2过期，3取消
	StartTime          int64          `json:"start_time" gorm:"bigint;not null"`                         // 开始时间
	EndTime            int64          `json:"end_time" gorm:"bigint;not null;index"`                     // 结束时间
	ModelQuotas        string         `json:"model_quotas" gorm:"type:text"`                             // 剩余模型配额JSON
	UsedQuotas         string         `json:"used_quotas" gorm:"type:text"`                              // 已使用配额JSON
	PurchasePrice      float64        `json:"purchase_price" gorm:"type:decimal(10,2)"`                  // 购买价格
	PaymentMethod      string         `json:"payment_method" gorm:"type:varchar(50)"`                    // 支付方式
	PaymentId          string         `json:"payment_id" gorm:"type:varchar(100);index"`                 // 支付订单ID
	CreatedTime        int64          `json:"created_time" gorm:"bigint;autoCreateTime"`                 // 创建时间
	UpdatedTime        int64          `json:"updated_time" gorm:"bigint;autoUpdateTime"`                 // 更新时间
	DeletedAt          gorm.DeletedAt `gorm:"index"`                                                     // 软删除
	
	// 关联字段
	User             *User             `json:"user,omitempty" gorm:"foreignKey:UserId"`
	SubscriptionPlan *SubscriptionPlan `json:"subscription_plan,omitempty" gorm:"foreignKey:SubscriptionPlanId"`
}

// 订阅状态常量
const (
	SubscriptionStatusActive   = 1 // 激活
	SubscriptionStatusExpired  = 2 // 过期
	SubscriptionStatusCanceled = 3 // 取消
)

// ModelQuotaInfo 模型配额信息
type ModelQuotaInfo struct {
	Total     int `json:"total"`     // 总配额
	Used      int `json:"used"`      // 已使用
	Remaining int `json:"remaining"` // 剩余
}

// GetModelQuotasMap 获取剩余模型配额映射
func (us *UserSubscription) GetModelQuotasMap() (ModelQuotaMap, error) {
	if us.ModelQuotas == "" {
		return make(ModelQuotaMap), nil
	}
	
	var quotas ModelQuotaMap
	err := json.Unmarshal([]byte(us.ModelQuotas), &quotas)
	if err != nil {
		return nil, fmt.Errorf("解析剩余配额失败: %v", err)
	}
	return quotas, nil
}

// SetModelQuotasMap 设置剩余模型配额映射
func (us *UserSubscription) SetModelQuotasMap(quotas ModelQuotaMap) error {
	data, err := json.Marshal(quotas)
	if err != nil {
		return fmt.Errorf("序列化剩余配额失败: %v", err)
	}
	us.ModelQuotas = string(data)
	return nil
}

// GetUsedQuotasMap 获取已使用配额映射
func (us *UserSubscription) GetUsedQuotasMap() (ModelQuotaMap, error) {
	if us.UsedQuotas == "" {
		return make(ModelQuotaMap), nil
	}
	
	var quotas ModelQuotaMap
	err := json.Unmarshal([]byte(us.UsedQuotas), &quotas)
	if err != nil {
		return nil, fmt.Errorf("解析已使用配额失败: %v", err)
	}
	return quotas, nil
}

// SetUsedQuotasMap 设置已使用配额映射
func (us *UserSubscription) SetUsedQuotasMap(quotas ModelQuotaMap) error {
	data, err := json.Marshal(quotas)
	if err != nil {
		return fmt.Errorf("序列化已使用配额失败: %v", err)
	}
	us.UsedQuotas = string(data)
	return nil
}

// Insert 创建用户订阅
func (us *UserSubscription) Insert() error {
	if us.UserId == 0 {
		return errors.New("用户ID不能为空")
	}
	if us.SubscriptionPlanId == 0 {
		return errors.New("订阅套餐ID不能为空")
	}
	
	us.CreatedTime = time.Now().Unix()
	us.UpdatedTime = us.CreatedTime
	return DB.Create(us).Error
}

// Update 更新用户订阅
func (us *UserSubscription) Update() error {
	if us.Id == 0 {
		return errors.New("订阅ID不能为空")
	}
	
	us.UpdatedTime = time.Now().Unix()
	return DB.Model(us).Updates(map[string]interface{}{
		"status":         us.Status,
		"model_quotas":   us.ModelQuotas,
		"used_quotas":    us.UsedQuotas,
		"updated_time":   us.UpdatedTime,
	}).Error
}

// GetUserSubscriptionById 根据ID获取用户订阅
func GetUserSubscriptionById(id int) (*UserSubscription, error) {
	if id == 0 {
		return nil, errors.New("无效的订阅ID")
	}
	
	var subscription UserSubscription
	err := DB.Preload("User").Preload("SubscriptionPlan").First(&subscription, id).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

// GetActiveUserSubscriptions 获取用户的激活订阅
func GetActiveUserSubscriptions(userId int) ([]*UserSubscription, error) {
	if userId == 0 {
		return nil, errors.New("无效的用户ID")
	}
	
	var subscriptions []*UserSubscription
	now := time.Now().Unix()
	
	err := DB.Where("user_id = ? AND status = ? AND start_time <= ? AND end_time > ?", 
		userId, SubscriptionStatusActive, now, now).
		Preload("SubscriptionPlan").
		Order("end_time ASC").
		Find(&subscriptions).Error
	
	return subscriptions, err
}

// GetUserSubscriptionsByPage 分页获取用户订阅
func GetUserSubscriptionsByPage(userId int, page, pageSize int) ([]*UserSubscription, int64, error) {
	var subscriptions []*UserSubscription
	var total int64
	
	query := DB.Model(&UserSubscription{})
	if userId > 0 {
		query = query.Where("user_id = ?", userId)
	}
	
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * pageSize
	err = query.Preload("User").Preload("SubscriptionPlan").
		Order("created_time DESC").
		Offset(offset).Limit(pageSize).
		Find(&subscriptions).Error
	
	return subscriptions, total, err
}

// ConsumeModelQuota 消费模型配额
func (us *UserSubscription) ConsumeModelQuota(modelName string, count int) error {
	if count <= 0 {
		return errors.New("消费数量必须大于0")
	}
	
	// 检查订阅是否激活
	if !us.IsActive() {
		return errors.New("订阅未激活或已过期")
	}
	
	// 获取剩余配额
	remainingQuotas, err := us.GetModelQuotasMap()
	if err != nil {
		return err
	}
	
	// 检查配额是否足够
	remaining, exists := remainingQuotas[modelName]
	if !exists || remaining < count {
		return fmt.Errorf("模型 %s 配额不足，剩余: %d，需要: %d", modelName, remaining, count)
	}
	
	// 更新剩余配额
	remainingQuotas[modelName] = remaining - count
	err = us.SetModelQuotasMap(remainingQuotas)
	if err != nil {
		return err
	}
	
	// 更新已使用配额
	usedQuotas, err := us.GetUsedQuotasMap()
	if err != nil {
		return err
	}
	
	if usedQuotas == nil {
		usedQuotas = make(ModelQuotaMap)
	}
	
	usedQuotas[modelName] += count
	err = us.SetUsedQuotasMap(usedQuotas)
	if err != nil {
		return err
	}
	
	// 保存到数据库
	return us.Update()
}

// GetModelQuotaInfo 获取指定模型的配额信息
func (us *UserSubscription) GetModelQuotaInfo(modelName string) (*ModelQuotaInfo, error) {
	// 获取套餐信息
	if us.SubscriptionPlan == nil {
		plan, err := GetSubscriptionPlanById(us.SubscriptionPlanId)
		if err != nil {
			return nil, err
		}
		us.SubscriptionPlan = plan
	}
	
	// 获取总配额
	total, err := us.SubscriptionPlan.GetModelQuotaByName(modelName)
	if err != nil {
		return nil, err
	}
	
	// 获取已使用配额
	usedQuotas, err := us.GetUsedQuotasMap()
	if err != nil {
		return nil, err
	}
	
	used := usedQuotas[modelName]
	remaining := total - used
	if remaining < 0 {
		remaining = 0
	}
	
	return &ModelQuotaInfo{
		Total:     total,
		Used:      used,
		Remaining: remaining,
	}, nil
}

// IsActive 检查订阅是否激活
func (us *UserSubscription) IsActive() bool {
	now := time.Now().Unix()
	return us.Status == SubscriptionStatusActive && 
		   us.StartTime <= now && 
		   us.EndTime > now
}

// IsExpired 检查订阅是否过期
func (us *UserSubscription) IsExpired() bool {
	now := time.Now().Unix()
	return us.EndTime <= now
}

// GetRemainingDays 获取剩余天数
func (us *UserSubscription) GetRemainingDays() int {
	now := time.Now().Unix()
	if us.EndTime <= now {
		return 0
	}
	
	return int((us.EndTime - now) / 86400) // 86400秒 = 1天
}

// CheckAndUpdateExpiredStatus 检查并更新过期状态
func (us *UserSubscription) CheckAndUpdateExpiredStatus() error {
	if us.IsExpired() && us.Status == SubscriptionStatusActive {
		us.Status = SubscriptionStatusExpired
		return us.Update()
	}
	return nil
}

// GetStatusText 获取状态文本
func (us *UserSubscription) GetStatusText() string {
	switch us.Status {
	case SubscriptionStatusActive:
		if us.IsExpired() {
			return "已过期"
		}
		return "激活"
	case SubscriptionStatusExpired:
		return "已过期"
	case SubscriptionStatusCanceled:
		return "已取消"
	default:
		return "未知"
	}
}

// CreateUserSubscription 创建用户订阅
func CreateUserSubscription(userId, planId int, paymentMethod, paymentId string) (*UserSubscription, error) {
	// 获取套餐信息
	plan, err := GetSubscriptionPlanById(planId)
	if err != nil {
		return nil, fmt.Errorf("获取套餐信息失败: %v", err)
	}
	
	if !plan.IsActive() {
		return nil, errors.New("套餐未启用")
	}
	
	// 获取套餐配额
	planQuotas, err := plan.GetModelQuotasMap()
	if err != nil {
		return nil, fmt.Errorf("获取套餐配额失败: %v", err)
	}
	
	now := time.Now().Unix()
	endTime := now + int64(plan.Duration*24*3600) // 转换为秒
	
	subscription := &UserSubscription{
		UserId:             userId,
		SubscriptionPlanId: planId,
		Status:             SubscriptionStatusActive,
		StartTime:          now,
		EndTime:            endTime,
		PurchasePrice:      plan.Price,
		PaymentMethod:      paymentMethod,
		PaymentId:          paymentId,
	}
	
	// 设置初始配额（剩余配额等于套餐配额）
	err = subscription.SetModelQuotasMap(planQuotas)
	if err != nil {
		return nil, fmt.Errorf("设置配额失败: %v", err)
	}
	
	// 初始化已使用配额为空
	err = subscription.SetUsedQuotasMap(make(ModelQuotaMap))
	if err != nil {
		return nil, fmt.Errorf("初始化已使用配额失败: %v", err)
	}
	
	err = subscription.Insert()
	if err != nil {
		return nil, fmt.Errorf("创建订阅失败: %v", err)
	}
	
	return subscription, nil
}
