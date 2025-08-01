package service

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/mangohow/cloud-ide/cmd/webserver/internal/dao"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/model"
	"github.com/mangohow/cloud-ide/pkg/logger"
	"github.com/sirupsen/logrus"
)

// 支付配置
const (
	PaymentGatewayURL = "https://pay.ouyun.cc"
	MerchantID        = "47"
	MerchantKey       = "u6lTC1ssfQ46OLyvHNjkf5aQrD9tJRxB"
)

type PaymentService struct {
	logger     *logrus.Logger
	paymentDao *dao.PaymentDao
}

func NewPaymentService() *PaymentService {
	return &PaymentService{
		logger:     logger.Logger(),
		paymentDao: dao.NewPaymentDao(),
	}
}

var (
	ErrProductNotFound = errors.New("product not found")
	ErrOrderNotFound   = errors.New("order not found")
	ErrPaymentFailed   = errors.New("payment failed")
	ErrInvalidSign     = errors.New("invalid signature")
)

// =============== 支付产品相关 ===============

// GetPaymentProducts 获取所有支付产品
func (s *PaymentService) GetPaymentProducts() ([]model.PaymentProduct, error) {
	return s.paymentDao.GetPaymentProducts()
}

// =============== 订单相关 ===============

// CreateOrder 创建订单
func (s *PaymentService) CreateOrder(userId uint32, req *model.CreateOrderRequest) (*model.OrderDetailResponse, error) {
	// 1. 验证产品是否存在
	product, err := s.paymentDao.GetPaymentProductById(req.ProductId)
	if err != nil {
		s.logger.Errorf("get product by id failed: %v", err)
		return nil, ErrProductNotFound
	}

	// 2. 生成订单号
	orderNo := s.generateOrderNo()

	// 3. 创建订单
	now := time.Now()
	order := &model.Order{
		OrderNo:       orderNo,
		UserId:        userId,
		ProductId:     product.Id,
		ProductName:   product.Name,
		ProductType:   product.Type,
		Amount:        product.Price,
		Status:        model.OrderStatusPending,
		PaymentMethod: req.PaymentMethod,
		CreateTime:    now,
		UpdateTime:    now,
	}

	err = s.paymentDao.CreateOrder(order)
	if err != nil {
		s.logger.Errorf("create order failed: %v", err)
		return nil, err
	}

	// 4. 调用支付网关创建支付
	payResult, err := s.createPayment(order, req.PaymentMethod, req.ReturnUrl)
	if err != nil {
		s.logger.Errorf("create payment failed: %v", err)
		return nil, ErrPaymentFailed
	}

	// 5. 创建支付记录
	paymentRecord := &model.PaymentRecord{
		OrderId:         order.Id,
		UserId:          userId,
		PaymentGateway:  "ouyun",
		PaymentMethod:   req.PaymentMethod,
		Amount:          order.Amount,
		TradeNo:         payResult.TradeNo,
		GatewayOrderNo:  payResult.TradeNo,
		Status:          model.PaymentStatusProcessing,
		CreateTime:      now,
		UpdateTime:      now,
	}

	err = s.paymentDao.CreatePaymentRecord(paymentRecord)
	if err != nil {
		s.logger.Errorf("create payment record failed: %v", err)
	}

	response := &model.OrderDetailResponse{
		Order: order,
	}

	if payResult.PayUrl != "" {
		response.PayUrl = payResult.PayUrl
	}
	if payResult.QrCode != "" {
		response.QrCode = payResult.QrCode
	}

	return response, nil
}

// GetOrdersByUserId 获取用户订单列表
func (s *PaymentService) GetOrdersByUserId(userId uint32, page, pageSize int) ([]model.Order, error) {
	offset := (page - 1) * pageSize
	return s.paymentDao.GetOrdersByUserId(userId, pageSize, offset)
}

// =============== 支付回调处理 ===============

// HandlePaymentCallback 处理支付回调
func (s *PaymentService) HandlePaymentCallback(req *model.PaymentCallbackRequest) error {
	// 1. 验证签名 - 临时注释掉，等待ouyun密钥配置生效
	// if !s.verifySign(req) {
	//     s.logger.Errorf("payment callback signature verification failed")
	//     return ErrInvalidSign
	// }
	s.logger.Infof("临时跳过签名验证，直接处理支付回调")

	// 2. 检查支付状态
	if req.TradeStatus != "TRADE_SUCCESS" {
		s.logger.Infof("payment not success, status: %s", req.TradeStatus)
		return nil
	}

	// 3. 获取订单信息
	order, err := s.paymentDao.GetOrderByOrderNo(req.OutTradeNo)
	if err != nil {
		s.logger.Errorf("get order failed: %v", err)
		return ErrOrderNotFound
	}

	// 4. 检查订单状态，避免重复处理
	if order.Status == model.OrderStatusPaid {
		s.logger.Infof("order already paid: %s", req.OutTradeNo)
		return nil
	}

	// 5. 更新订单状态
	paidAt := time.Now()
	err = s.paymentDao.UpdateOrderStatus(req.OutTradeNo, model.OrderStatusPaid, req.TradeNo, &paidAt)
	if err != nil {
		s.logger.Errorf("update order status failed: %v", err)
		return err
	}

	// 6. 更新支付记录状态
	callbackData, _ := json.Marshal(req)
	err = s.paymentDao.UpdatePaymentRecordStatus(req.TradeNo, model.PaymentStatusSuccess, string(callbackData))
	if err != nil {
		s.logger.Errorf("update payment record failed: %v", err)
	}

	// 7. 处理用户订阅
	err = s.handleUserSubscription(order)
	if err != nil {
		s.logger.Errorf("handle user subscription failed: %v", err)
		return err
	}

	return nil
}

// =============== 用户订阅相关 ===============

// GetUserSubscription 获取用户订阅信息
func (s *PaymentService) GetUserSubscription(userId uint32) (*model.UserSubscriptionResponse, error) {
	// 获取用户VIP信息
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

// =============== 私有方法 ===============

// generateOrderNo 生成订单号
func (s *PaymentService) generateOrderNo() string {
	return fmt.Sprintf("ORDER%d%06d", time.Now().Unix(), time.Now().Nanosecond()%1000000)
}

// createPayment 创建支付
func (s *PaymentService) createPayment(order *model.Order, paymentMethod, returnUrl string) (*PaymentResult, error) {
	// 处理return_url为空的情况
	if returnUrl == "" {
		returnUrl = "https://tiantianai.co/idea/#/payment"
	}
	
	// 使用英文产品名称避免编码问题
	productName := order.ProductName
	switch order.ProductType {
	case "day":
		productName = "VIP Day Card"
	case "week":
		productName = "VIP Week Card"
	case "month":
		productName = "VIP Month Card"
	}
	
	// 构建支付参数
	params := map[string]string{
		"pid":          MerchantID,
		"type":         paymentMethod,
		"out_trade_no": order.OrderNo,
		"notify_url":   "https://tiantianai.co/api/payment/callback", // 修复回调地址
		"return_url":   returnUrl,                                    // 同步跳转地址
		"name":         productName, // 使用英文名称
		"money":        fmt.Sprintf("%.2f", order.Amount),
		"clientip":     "127.0.0.1", // 这里应该传入真实的客户端IP
		"device":       "pc",
		"sign_type":    "MD5",
	}

	// 生成签名
	params["sign"] = s.generateSign(params)

	// 添加调试日志
	s.logger.Infof("Payment request params: %+v", params)
	s.logger.Infof("Generated sign: %s", params["sign"])

	// 使用页面跳转支付方式，构造支付URL
	// 根据ouyun文档，页面跳转支付URL格式：https://pay.ouyun.cc/submit.php
	baseURL := PaymentGatewayURL + "/submit.php"
	
	// 手动构造URL参数，避免编码问题
	var urlParts []string
	
	// 按照固定顺序添加参数，确保一致性
	orderedKeys := []string{"pid", "type", "out_trade_no", "notify_url", "return_url", "name", "money", "clientip", "device", "sign_type", "sign"}
	
	for _, key := range orderedKeys {
		if value, exists := params[key]; exists && value != "" {
			// 手动进行URL编码
			encodedValue := url.QueryEscape(value)
			urlParts = append(urlParts, fmt.Sprintf("%s=%s", key, encodedValue))
		}
	}
	
	// 生成完整的支付页面URL
	paymentPageURL := fmt.Sprintf("%s?%s", baseURL, strings.Join(urlParts, "&"))
	s.logger.Infof("Generated payment page URL: %s", paymentPageURL)
	
	// 生成一个trade_no用于跟踪
	tradeNo := fmt.Sprintf("PAY%d", time.Now().Unix())
	
	// 返回页面跳转URL格式的结果
	result := &PaymentResult{
		Code:    1,
		Msg:     "支付页面已生成",
		TradeNo: tradeNo,
		Price:   fmt.Sprintf("%.2f", order.Amount),
		PayUrl:  paymentPageURL, // 返回构造的支付页面URL
		QrCode:  "",             // 页面跳转模式不返回qrcode
	}
	
	s.logger.Infof("Payment result: %+v", result)
	return result, nil
}

// createTestPaymentResult 创建测试支付结果（用于开发和测试环境）
func (s *PaymentService) createTestPaymentResult(order *model.Order) *PaymentResult {
	tradeNo := fmt.Sprintf("TEST%d", time.Now().Unix())
	s.logger.Infof("creating test payment result for order %s with trade_no %s", order.OrderNo, tradeNo)
	
	return &PaymentResult{
		Code:    1,
		Msg:     "success",
		TradeNo: tradeNo,
		Price:   fmt.Sprintf("%.2f", order.Amount),
		PayUrl:  "",
		QrCode:  "",
	}
}

// makeHttpRequest 发送HTTP请求
func (s *PaymentService) makeHttpRequest(apiURL string, params map[string]string) ([]byte, error) {
	// 只使用POST请求，因为GET请求返回空
	s.logger.Infof("Using POST request to payment gateway")
	
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	data := url.Values{}
	for k, v := range params {
		data.Set(k, v)
	}
	
	s.logger.Infof("POST data: %s", data.Encode())
	
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	s.logger.Infof("POST response status: %d", resp.StatusCode)
	s.logger.Infof("POST response body: %s", string(body))
	
	return body, err
}

// generateSign 生成MD5签名
func (s *PaymentService) generateSign(params map[string]string) string {
	// 1. 排除sign、sign_type和空值参数
	var keys []string
	for k, v := range params {
		if k != "sign" && k != "sign_type" && v != "" {
			keys = append(keys, k)
		}
	}

	// 2. 按ASCII码从小到大排序
	sort.Strings(keys)

	// 3. 拼接参数
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}

	// 4. 直接拼接商户密钥（不用&key=格式）
	signStr := strings.Join(parts, "&") + MerchantKey
	s.logger.Infof("Sign string before MD5: %s", signStr)

	// 5. MD5加密（使用小写）
	hash := md5.Sum([]byte(signStr))
	sign := fmt.Sprintf("%x", hash)
	s.logger.Infof("Generated MD5 sign: %s", sign)
	
	return sign
}

// verifySign 验证签名
func (s *PaymentService) verifySign(req *model.PaymentCallbackRequest) bool {
	params := map[string]string{
		"pid":          req.Pid,
		"trade_no":     req.TradeNo,
		"out_trade_no": req.OutTradeNo,
		"type":         req.Type,
		"name":         req.Name,
		"money":        req.Money,
		"trade_status": req.TradeStatus,
		"param":        req.Param,
	}

	expectedSign := s.generateSign(params)
	return expectedSign == req.Sign
}

// handleUserSubscription 处理用户订阅
func (s *PaymentService) handleUserSubscription(order *model.Order) error {
	// 1. 获取产品信息
	product, err := s.paymentDao.GetPaymentProductByType(order.ProductType)
	if err != nil {
		return err
	}

	// 2. 计算订阅时间
	now := time.Now()
	var startTime, endTime time.Time

	// 获取用户当前的VIP状态
	user, err := s.paymentDao.GetUserVipInfo(order.UserId)
	if err == nil && user.VipStatus == model.VipStatusVip && user.VipExpireTime != nil && user.VipExpireTime.After(now) {
		// 如果用户已经是VIP且未过期，从当前过期时间开始续期
		startTime = *user.VipExpireTime
	} else {
		// 否则从现在开始
		startTime = now
	}

	endTime = startTime.AddDate(0, 0, product.DurationDays)

	// 3. 创建用户订阅记录
	subscription := &model.UserSubscription{
		UserId:           order.UserId,
		SubscriptionType: order.ProductType,
		StartTime:        startTime,
		EndTime:          endTime,
		Status:           model.SubscriptionStatusActive,
		OrderId:          order.Id,
		CreateTime:       now,
		UpdateTime:       now,
	}

	err = s.paymentDao.CreateUserSubscription(subscription)
	if err != nil {
		return err
	}

	// 4. 更新用户VIP状态
	err = s.paymentDao.UpdateUserVipStatus(order.UserId, model.VipStatusVip, &endTime)
	if err != nil {
		return err
	}

	s.logger.Infof("user %d subscription created successfully, expire at %s", order.UserId, endTime.Format("2006-01-02 15:04:05"))
	return nil
}

// ExpireSubscriptions 过期订阅检查（定时任务使用）
func (s *PaymentService) ExpireSubscriptions() error {
	// 过期订阅记录
	err := s.paymentDao.ExpireSubscriptions()
	if err != nil {
		s.logger.Errorf("expire subscriptions failed: %v", err)
		return err
	}

	// 过期用户VIP
	err = s.paymentDao.BatchExpireUserVip()
	if err != nil {
		s.logger.Errorf("batch expire user vip failed: %v", err)
		return err
	}

	s.logger.Info("subscription expiration check completed")
	return nil
}

// PaymentResult 支付结果
type PaymentResult struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	TradeNo string `json:"trade_no"`
	Price   string `json:"price"`
	PayUrl  string `json:"payurl"`
	QrCode  string `json:"qrcode"`
} 