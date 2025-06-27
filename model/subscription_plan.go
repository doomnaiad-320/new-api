package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"one-api/common"
	"time"

	"gorm.io/gorm"
)

// SubscriptionPlan 订阅套餐表
type SubscriptionPlan struct {
	Id          int            `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"type:varchar(100);not null;index"`                    // 套餐名称
	Description string         `json:"description" gorm:"type:text"`                                    // 套餐描述
	Price       float64        `json:"price" gorm:"type:decimal(10,2);not null"`                       // 套餐价格（元）
	Duration    int            `json:"duration" gorm:"default:30"`                                     // 有效期（天）
	Status      int            `json:"status" gorm:"default:1"`                                        // 状态：1启用，0禁用
	ModelQuotas string         `json:"model_quotas" gorm:"type:text"`                                  // 模型配额JSON，格式：{"gpt-4": 100, "claude-3": 50}
	CreatedTime int64          `json:"created_time" gorm:"bigint;autoCreateTime"`                      // 创建时间
	UpdatedTime int64          `json:"updated_time" gorm:"bigint;autoUpdateTime"`                      // 更新时间
	DeletedAt   gorm.DeletedAt `gorm:"index"`                                                          // 软删除
}

// ModelQuotaMap 模型配额映射
type ModelQuotaMap map[string]int

// GetModelQuotasMap 获取模型配额映射
func (sp *SubscriptionPlan) GetModelQuotasMap() (ModelQuotaMap, error) {
	if sp.ModelQuotas == "" {
		return make(ModelQuotaMap), nil
	}
	
	var quotas ModelQuotaMap
	err := json.Unmarshal([]byte(sp.ModelQuotas), &quotas)
	if err != nil {
		return nil, fmt.Errorf("解析模型配额失败: %v", err)
	}
	return quotas, nil
}

// SetModelQuotasMap 设置模型配额映射
func (sp *SubscriptionPlan) SetModelQuotasMap(quotas ModelQuotaMap) error {
	data, err := json.Marshal(quotas)
	if err != nil {
		return fmt.Errorf("序列化模型配额失败: %v", err)
	}
	sp.ModelQuotas = string(data)
	return nil
}

// Insert 创建订阅套餐
func (sp *SubscriptionPlan) Insert() error {
	if sp.Name == "" {
		return errors.New("套餐名称不能为空")
	}
	if sp.Price < 0 {
		return errors.New("套餐价格不能为负数")
	}
	if sp.Duration <= 0 {
		return errors.New("有效期必须大于0")
	}
	
	sp.CreatedTime = time.Now().Unix()
	sp.UpdatedTime = sp.CreatedTime
	return DB.Create(sp).Error
}

// Update 更新订阅套餐
func (sp *SubscriptionPlan) Update() error {
	if sp.Id == 0 {
		return errors.New("套餐ID不能为空")
	}
	
	sp.UpdatedTime = time.Now().Unix()
	return DB.Model(sp).Updates(map[string]interface{}{
		"name":         sp.Name,
		"description":  sp.Description,
		"price":        sp.Price,
		"duration":     sp.Duration,
		"status":       sp.Status,
		"model_quotas": sp.ModelQuotas,
		"updated_time": sp.UpdatedTime,
	}).Error
}

// Delete 删除订阅套餐（软删除）
func (sp *SubscriptionPlan) Delete() error {
	if sp.Id == 0 {
		return errors.New("套餐ID不能为空")
	}
	return DB.Delete(sp).Error
}

// GetSubscriptionPlanById 根据ID获取订阅套餐
func GetSubscriptionPlanById(id int) (*SubscriptionPlan, error) {
	if id == 0 {
		return nil, errors.New("无效的套餐ID")
	}
	
	var plan SubscriptionPlan
	err := DB.First(&plan, id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// GetAllSubscriptionPlans 获取所有启用的订阅套餐
func GetAllSubscriptionPlans(status int) ([]*SubscriptionPlan, error) {
	var plans []*SubscriptionPlan
	query := DB.Order("price ASC")
	
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	
	err := query.Find(&plans).Error
	return plans, err
}

// GetSubscriptionPlansByPage 分页获取订阅套餐
func GetSubscriptionPlansByPage(page, pageSize int, status int) ([]*SubscriptionPlan, int64, error) {
	var plans []*SubscriptionPlan
	var total int64
	
	query := DB.Model(&SubscriptionPlan{})
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * pageSize
	err = query.Order("created_time DESC").Offset(offset).Limit(pageSize).Find(&plans).Error
	return plans, total, err
}

// SearchSubscriptionPlans 搜索订阅套餐
func SearchSubscriptionPlans(keyword string, status int) ([]*SubscriptionPlan, error) {
	var plans []*SubscriptionPlan
	query := DB.Model(&SubscriptionPlan{})
	
	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	
	err := query.Order("created_time DESC").Find(&plans).Error
	return plans, err
}

// ValidateModelQuotas 验证模型配额格式
func ValidateModelQuotas(modelQuotas string) error {
	if modelQuotas == "" {
		return nil
	}
	
	var quotas ModelQuotaMap
	err := json.Unmarshal([]byte(modelQuotas), &quotas)
	if err != nil {
		return fmt.Errorf("模型配额格式错误: %v", err)
	}
	
	for model, quota := range quotas {
		if model == "" {
			return errors.New("模型名称不能为空")
		}
		if quota < 0 {
			return fmt.Errorf("模型 %s 的配额不能为负数", model)
		}
	}
	
	return nil
}

// GetModelQuotaByName 获取指定模型的配额
func (sp *SubscriptionPlan) GetModelQuotaByName(modelName string) (int, error) {
	quotas, err := sp.GetModelQuotasMap()
	if err != nil {
		return 0, err
	}
	
	quota, exists := quotas[modelName]
	if !exists {
		return 0, nil // 如果模型不在套餐中，返回0配额
	}
	
	return quota, nil
}

// HasModel 检查套餐是否包含指定模型
func (sp *SubscriptionPlan) HasModel(modelName string) bool {
	quotas, err := sp.GetModelQuotasMap()
	if err != nil {
		return false
	}
	
	_, exists := quotas[modelName]
	return exists
}

// GetTotalQuota 获取套餐总配额数
func (sp *SubscriptionPlan) GetTotalQuota() (int, error) {
	quotas, err := sp.GetModelQuotasMap()
	if err != nil {
		return 0, err
	}
	
	total := 0
	for _, quota := range quotas {
		total += quota
	}
	
	return total, nil
}

// IsActive 检查套餐是否启用
func (sp *SubscriptionPlan) IsActive() bool {
	return sp.Status == 1
}

// GetFormattedPrice 获取格式化的价格字符串
func (sp *SubscriptionPlan) GetFormattedPrice() string {
	return fmt.Sprintf("%.2f", sp.Price)
}

// GetFormattedDuration 获取格式化的有效期字符串
func (sp *SubscriptionPlan) GetFormattedDuration() string {
	if sp.Duration == 1 {
		return "1天"
	} else if sp.Duration == 7 {
		return "1周"
	} else if sp.Duration == 30 {
		return "1个月"
	} else if sp.Duration == 90 {
		return "3个月"
	} else if sp.Duration == 365 {
		return "1年"
	}
	return fmt.Sprintf("%d天", sp.Duration)
}
