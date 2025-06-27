package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// SubscriptionPlanRequest 订阅套餐请求结构
type SubscriptionPlanRequest struct {
	Name        string                   `json:"name" binding:"required"`
	Description string                   `json:"description"`
	Price       float64                  `json:"price" binding:"required,min=0"`
	Duration    int                      `json:"duration" binding:"required,min=1"`
	Status      int                      `json:"status"`
	ModelQuotas model.ModelQuotaMap      `json:"model_quotas" binding:"required"`
}

// PurchaseSubscriptionRequest 购买订阅请求结构
type PurchaseSubscriptionRequest struct {
	PlanId        int    `json:"plan_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
}

// GetAllSubscriptionPlans 获取所有订阅套餐
func GetAllSubscriptionPlans(c *gin.Context) {
	status := -1 // 默认获取所有状态
	if statusStr := c.Query("status"); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			status = s
		}
	}
	
	plans, err := model.GetAllSubscriptionPlans(status)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取订阅套餐失败: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    plans,
	})
}

// GetSubscriptionPlansByPage 分页获取订阅套餐
func GetSubscriptionPlansByPage(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := -1
	
	if statusStr := c.Query("status"); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			status = s
		}
	}
	
	plans, total, err := model.GetSubscriptionPlansByPage(page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取订阅套餐失败: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"plans":     plans,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetSubscriptionPlan 获取单个订阅套餐
func GetSubscriptionPlan(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的套餐ID",
		})
		return
	}
	
	plan, err := model.GetSubscriptionPlanById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取订阅套餐失败: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    plan,
	})
}

// CreateSubscriptionPlan 创建订阅套餐（管理员）
func CreateSubscriptionPlan(c *gin.Context) {
	var req SubscriptionPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}
	
	// 验证模型配额格式
	modelQuotasJSON, err := json.Marshal(req.ModelQuotas)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "模型配额格式错误",
		})
		return
	}
	
	plan := &model.SubscriptionPlan{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Duration:    req.Duration,
		Status:      req.Status,
		ModelQuotas: string(modelQuotasJSON),
	}
	
	err = plan.Insert()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "创建订阅套餐失败: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "创建订阅套餐成功",
		"data":    plan,
	})
}

// UpdateSubscriptionPlan 更新订阅套餐（管理员）
func UpdateSubscriptionPlan(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的套餐ID",
		})
		return
	}
	
	var req SubscriptionPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}
	
	plan, err := model.GetSubscriptionPlanById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取订阅套餐失败: " + err.Error(),
		})
		return
	}
	
	// 验证模型配额格式
	modelQuotasJSON, err := json.Marshal(req.ModelQuotas)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "模型配额格式错误",
		})
		return
	}
	
	plan.Name = req.Name
	plan.Description = req.Description
	plan.Price = req.Price
	plan.Duration = req.Duration
	plan.Status = req.Status
	plan.ModelQuotas = string(modelQuotasJSON)
	
	err = plan.Update()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "更新订阅套餐失败: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "更新订阅套餐成功",
		"data":    plan,
	})
}

// DeleteSubscriptionPlan 删除订阅套餐（管理员）
func DeleteSubscriptionPlan(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的套餐ID",
		})
		return
	}
	
	plan, err := model.GetSubscriptionPlanById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取订阅套餐失败: " + err.Error(),
		})
		return
	}
	
	err = plan.Delete()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "删除订阅套餐失败: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "删除订阅套餐成功",
	})
}

// PurchaseSubscription 购买订阅
func PurchaseSubscription(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户未登录",
		})
		return
	}
	
	var req PurchaseSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}
	
	// 检查套餐是否存在且启用
	plan, err := model.GetSubscriptionPlanById(req.PlanId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取订阅套餐失败: " + err.Error(),
		})
		return
	}
	
	if !plan.IsActive() {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "订阅套餐未启用",
		})
		return
	}
	
	// 生成支付订单ID（这里简化处理，实际应该集成支付系统）
	paymentId := fmt.Sprintf("SUB_%d_%d_%d", userId, req.PlanId, time.Now().Unix())
	
	// 创建用户订阅
	subscription, err := model.CreateUserSubscription(userId, req.PlanId, req.PaymentMethod, paymentId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "创建订阅失败: " + err.Error(),
		})
		return
	}
	
	// 记录日志
	model.RecordLog(userId, model.LogTypeSystem, fmt.Sprintf("购买订阅套餐: %s，价格: %.2f元", plan.Name, plan.Price))
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "购买订阅成功",
		"data": gin.H{
			"subscription": subscription,
			"payment_id":   paymentId,
		},
	})
}

// GetUserSubscriptions 获取用户订阅列表
func GetUserSubscriptions(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户未登录",
		})
		return
	}
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	
	subscriptions, total, err := model.GetUserSubscriptionsByPage(userId, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取用户订阅失败: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"subscriptions": subscriptions,
			"total":         total,
			"page":          page,
			"page_size":     pageSize,
		},
	})
}

// GetActiveUserSubscriptions 获取用户激活的订阅
func GetActiveUserSubscriptions(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户未登录",
		})
		return
	}
	
	subscriptions, err := model.GetActiveUserSubscriptions(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取激活订阅失败: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    subscriptions,
	})
}

// GetSubscriptionQuotas 获取用户订阅配额信息
func GetSubscriptionQuotas(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户未登录",
		})
		return
	}

	subscriptions, err := model.GetActiveUserSubscriptions(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取激活订阅失败: " + err.Error(),
		})
		return
	}

	// 汇总所有激活订阅的配额
	totalQuotas := make(map[string]*model.ModelQuotaInfo)

	for _, subscription := range subscriptions {
		quotas, err := subscription.GetModelQuotasMap()
		if err != nil {
			continue
		}

		for modelName, remaining := range quotas {
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

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"quotas":       totalQuotas,
			"subscription_count": len(subscriptions),
		},
	})
}

// GetSubscriptionUsage 获取订阅使用记录
func GetSubscriptionUsage(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户未登录",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	usages, total, err := model.GetUsageByUser(userId, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取使用记录失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"usages":    usages,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetSubscriptionStats 获取订阅统计信息
func GetSubscriptionStats(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户未登录",
		})
		return
	}

	// 获取时间范围参数
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days <= 0 || days > 365 {
		days = 30
	}

	// 计算时间范围
	endTime := time.Now().Unix()
	startTime := endTime - int64(days*24*3600)

	// 获取使用统计
	stats, err := model.GetUsageStatsByUser(userId, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取统计信息失败: " + err.Error(),
		})
		return
	}

	// 获取每日使用统计
	dailyStats, err := model.GetDailyUsageStats(userId, days)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取每日统计失败: " + err.Error(),
		})
		return
	}

	// 获取使用量最高的模型
	topModels, err := model.GetTopModelsUsage(userId, 10, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取热门模型失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"period_stats": stats,
			"daily_stats":  dailyStats,
			"top_models":   topModels,
			"days":         days,
		},
	})
}

// SearchSubscriptionPlans 搜索订阅套餐
func SearchSubscriptionPlans(c *gin.Context) {
	keyword := c.Query("keyword")
	status := -1

	if statusStr := c.Query("status"); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			status = s
		}
	}

	plans, err := model.SearchSubscriptionPlans(keyword, status)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "搜索订阅套餐失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    plans,
	})
}

// GetAllUserSubscriptions 获取所有用户订阅（管理员）
func GetAllUserSubscriptions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	subscriptions, total, err := model.GetUserSubscriptionsByPage(0, page, pageSize) // userId=0表示获取所有用户
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取用户订阅失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"subscriptions": subscriptions,
			"total":         total,
			"page":          page,
			"page_size":     pageSize,
		},
	})
}

// GetSubscriptionReport 获取订阅报表（管理员）
func GetSubscriptionReport(c *gin.Context) {
	startTime, _ := strconv.ParseInt(c.DefaultQuery("start_time", "0"), 10, 64)
	endTime, _ := strconv.ParseInt(c.DefaultQuery("end_time", "0"), 10, 64)

	monitorService := service.NewSubscriptionMonitorService()
	report, err := monitorService.GenerateSubscriptionReport(startTime, endTime)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "生成订阅报表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    report,
	})
}

// GetSystemSubscriptionStats 获取系统订阅统计（管理员）
func GetSystemSubscriptionStats(c *gin.Context) {
	monitorService := service.NewSubscriptionMonitorService()
	stats, err := monitorService.GetSystemSubscriptionStats()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取系统订阅统计失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetUserSubscriptionSummary 获取用户订阅摘要
func GetUserSubscriptionSummary(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户未登录",
		})
		return
	}

	monitorService := service.NewSubscriptionMonitorService()
	summary, err := monitorService.GetUserSubscriptionSummary(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取用户订阅摘要失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    summary,
	})
}

// TriggerSubscriptionMonitor 手动触发订阅监控（管理员）
func TriggerSubscriptionMonitor(c *gin.Context) {
	monitorService := service.NewSubscriptionMonitorService()

	err := monitorService.MonitorSubscriptionQuotas()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "执行订阅监控失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "订阅监控执行成功",
	})
}

// CleanupExpiredSubscriptions 清理过期订阅（管理员）
func CleanupExpiredSubscriptions(c *gin.Context) {
	monitorService := service.NewSubscriptionMonitorService()

	err := monitorService.CleanupExpiredSubscriptions()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "清理过期订阅失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "过期订阅清理成功",
	})
}
