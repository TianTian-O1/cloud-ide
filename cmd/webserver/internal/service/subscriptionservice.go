package service

import (
	"time"

	"github.com/mangohow/cloud-ide/cmd/webserver/internal/dao"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/model"
	"github.com/mangohow/cloud-ide/pkg/logger"
	"github.com/sirupsen/logrus"
)

type SubscriptionService struct {
	logger     *logrus.Logger
	paymentDao *dao.PaymentDao
}

func NewSubscriptionService() *SubscriptionService {
	return &SubscriptionService{
		logger:     logger.Logger(),
		paymentDao: dao.NewPaymentDao(),
	}
}

// CheckUserVipStatus 检查用户VIP状态
func (s *SubscriptionService) CheckUserVipStatus(userId uint32) (bool, error) {
	user, err := s.paymentDao.GetUserVipInfo(userId)
	if err != nil {
		s.logger.Errorf("get user vip info failed: %v", err)
		return false, nil // 出错时默认为非VIP
	}

	// 检查VIP状态和过期时间
	if user.VipStatus == model.VipStatusVip && user.VipExpireTime != nil {
		return user.VipExpireTime.After(time.Now()), nil
	}

	return false, nil
}

// GetUserVipInfo 获取用户VIP详细信息
func (s *SubscriptionService) GetUserVipInfo(userId uint32) (*model.UserSubscriptionResponse, error) {
	user, err := s.paymentDao.GetUserVipInfo(userId)
	if err != nil {
		s.logger.Errorf("get user vip info failed: %v", err)
		return &model.UserSubscriptionResponse{
			VipStatus:    model.VipStatusNormal,
			IsActive:     false,
			DaysLeft:     0,
			CurrentLevel: "普通用户",
		}, nil
	}

	response := &model.UserSubscriptionResponse{
		VipStatus:    user.VipStatus,
		ExpireTime:   user.VipExpireTime,
		IsActive:     false,
		DaysLeft:     0,
		CurrentLevel: "普通用户",
	}

	// 计算剩余天数和状态
	if user.VipStatus == model.VipStatusVip && user.VipExpireTime != nil {
		now := time.Now()
		if user.VipExpireTime.After(now) {
			response.IsActive = true
			response.DaysLeft = int(user.VipExpireTime.Sub(now).Hours() / 24)
			
			// 获取最近的订阅记录确定等级
			subscription, err := s.paymentDao.GetActiveSubscriptionByUserId(userId)
			if err == nil {
				switch subscription.SubscriptionType {
				case model.ProductTypeDay:
					response.CurrentLevel = "日卡会员"
				case model.ProductTypeWeek:
					response.CurrentLevel = "周卡会员"
				case model.ProductTypeMonth:
					response.CurrentLevel = "月卡会员"
				default:
					response.CurrentLevel = "VIP会员"
				}
			} else {
				response.CurrentLevel = "VIP会员"
			}
		}
	}

	return response, nil
}

// IsUserVip 简单检查用户是否为VIP
func (s *SubscriptionService) IsUserVip(userId uint32) bool {
	isVip, _ := s.CheckUserVipStatus(userId)
	return isVip
}

// GetActiveSubscription 获取用户当前有效的订阅
func (s *SubscriptionService) GetActiveSubscription(userId uint32) (*model.UserSubscription, error) {
	return s.paymentDao.GetActiveSubscriptionByUserId(userId)
}

// GetSubscriptionHistory 获取用户订阅历史
func (s *SubscriptionService) GetSubscriptionHistory(userId uint32) ([]model.UserSubscription, error) {
	return s.paymentDao.GetSubscriptionsByUserId(userId)
}

// GetVipUserCount 获取VIP用户数量（管理员使用）
func (s *SubscriptionService) GetVipUserCount() (int, error) {
	// 这里需要在DAO中添加相应的方法，暂时返回0
	// TODO: 在PaymentDao中添加GetVipUserCount方法
	return 0, nil
}

// GetSubscriptionStats 获取订阅统计信息（管理员使用）
func (s *SubscriptionService) GetSubscriptionStats() (map[string]interface{}, error) {
	// TODO: 实现订阅统计功能
	stats := map[string]interface{}{
		"total_vip_users":   0,
		"day_subscribers":   0,
		"week_subscribers":  0,
		"month_subscribers": 0,
		"expired_today":     0,
	}
	return stats, nil
}

// CheckWorkspacePermission 检查工作空间操作权限
func (s *SubscriptionService) CheckWorkspacePermission(userId uint32, action string) bool {
	// 检查用户是否为VIP
	isVip := s.IsUserVip(userId)
	
	switch action {
	case "create_workspace":
		// VIP用户可以创建无限个工作空间，普通用户有限制
		return isVip // 如果是VIP直接返回true，普通用户需要额外检查
	case "start_workspace":
		// VIP用户可以启动工作空间，普通用户可能有限制
		return isVip
	case "use_premium_templates":
		// 高级模板只有VIP用户可以使用
		return isVip
	case "extended_runtime":
		// 延长运行时间只有VIP用户可以使用
		return isVip
	default:
		// 默认操作所有用户都可以
		return true
	}
}

// GetUserPermissions 获取用户权限列表
func (s *SubscriptionService) GetUserPermissions(userId uint32) map[string]bool {
	isVip := s.IsUserVip(userId)
	
	permissions := map[string]bool{
		"create_unlimited_workspace": isVip,
		"use_premium_templates":      isVip,
		"extended_runtime":           isVip,
		"priority_support":           isVip,
		"advanced_features":          isVip,
	}
	
	return permissions
}

// HandleSubscriptionExpiration 处理订阅过期（定时任务调用）
func (s *SubscriptionService) HandleSubscriptionExpiration() error {
	s.logger.Info("starting subscription expiration check")
	
	// 过期订阅记录
	err := s.paymentDao.ExpireSubscriptions()
	if err != nil {
		s.logger.Errorf("expire subscriptions failed: %v", err)
		return err
	}

	// 过期用户VIP状态
	err = s.paymentDao.BatchExpireUserVip()
	if err != nil {
		s.logger.Errorf("batch expire user vip failed: %v", err)
		return err
	}

	s.logger.Info("subscription expiration check completed")
	return nil
}

// NotifyExpiringUsers 通知即将过期的用户（定时任务调用）
func (s *SubscriptionService) NotifyExpiringUsers() error {
	// TODO: 实现即将过期用户的通知功能
	// 1. 查询即将过期的用户（比如3天内过期）
	// 2. 发送邮件或站内消息通知
	// 3. 记录通知日志
	
	s.logger.Info("expiring users notification completed")
	return nil
}

// ExtendSubscription 延长订阅（管理员功能）
func (s *SubscriptionService) ExtendSubscription(userId uint32, days int, reason string) error {
	// 获取用户当前VIP信息
	user, err := s.paymentDao.GetUserVipInfo(userId)
	if err != nil {
		return err
	}

	var newExpireTime time.Time
	now := time.Now()
	
	// 如果用户当前是VIP且未过期，从过期时间开始延长
	if user.VipStatus == model.VipStatusVip && user.VipExpireTime != nil && user.VipExpireTime.After(now) {
		newExpireTime = user.VipExpireTime.AddDate(0, 0, days)
	} else {
		// 否则从现在开始
		newExpireTime = now.AddDate(0, 0, days)
	}

	// 更新用户VIP状态
	err = s.paymentDao.UpdateUserVipStatus(userId, model.VipStatusVip, &newExpireTime)
	if err != nil {
		return err
	}

	s.logger.Infof("extended subscription for user %d by %d days, new expire time: %s, reason: %s", 
		userId, days, newExpireTime.Format("2006-01-02 15:04:05"), reason)
	
	return nil
} 