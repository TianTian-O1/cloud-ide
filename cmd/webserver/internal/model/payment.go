package model

import (
	"time"
)

// =============== 支付产品相关 ===============

// PaymentProduct 支付产品表
type PaymentProduct struct {
	Id           uint32    `db:"id" json:"id"`
	Name         string    `db:"name" json:"name"`
	Type         string    `db:"type" json:"type"`
	DurationDays int       `db:"duration_days" json:"duration_days"`
	Price        float64   `db:"price" json:"price"`
	Description  string    `db:"description" json:"description"`
	Status       uint8     `db:"status" json:"status"`
	CreateTime   time.Time `db:"create_time" json:"create_time"`
	UpdateTime   time.Time `db:"update_time" json:"update_time"`
}

// =============== 订单相关 ===============

// Order 订单表
type Order struct {
	Id            uint32     `db:"id" json:"id"`
	OrderNo       string     `db:"order_no" json:"order_no"`
	UserId        uint32     `db:"user_id" json:"user_id"`
	ProductId     uint32     `db:"product_id" json:"product_id"`
	ProductName   string     `db:"product_name" json:"product_name"`
	ProductType   string     `db:"product_type" json:"product_type"`
	Amount        float64    `db:"amount" json:"amount"`
	Status        uint8      `db:"status" json:"status"`
	PaymentMethod string     `db:"payment_method" json:"payment_method"`
	TradeNo       string     `db:"trade_no" json:"trade_no"`
	PaidAt        *time.Time `db:"paid_at" json:"paid_at"`
	CreateTime    time.Time  `db:"create_time" json:"create_time"`
	UpdateTime    time.Time  `db:"update_time" json:"update_time"`
}

// =============== 用户订阅相关 ===============

// UserSubscription 用户订阅表
type UserSubscription struct {
	Id               uint32    `db:"id" json:"id"`
	UserId           uint32    `db:"user_id" json:"user_id"`
	SubscriptionType string    `db:"subscription_type" json:"subscription_type"`
	StartTime        time.Time `db:"start_time" json:"start_time"`
	EndTime          time.Time `db:"end_time" json:"end_time"`
	Status           uint8     `db:"status" json:"status"`
	OrderId          uint32    `db:"order_id" json:"order_id"`
	CreateTime       time.Time `db:"create_time" json:"create_time"`
	UpdateTime       time.Time `db:"update_time" json:"update_time"`
}

// =============== 支付记录相关 ===============

// PaymentRecord 支付记录表
type PaymentRecord struct {
	Id             uint32    `db:"id" json:"id"`
	OrderId        uint32    `db:"order_id" json:"order_id"`
	UserId         uint32    `db:"user_id" json:"user_id"`
	PaymentGateway string    `db:"payment_gateway" json:"payment_gateway"`
	PaymentMethod  string    `db:"payment_method" json:"payment_method"`
	Amount         float64   `db:"amount" json:"amount"`
	TradeNo        string    `db:"trade_no" json:"trade_no"`
	GatewayOrderNo string    `db:"gateway_order_no" json:"gateway_order_no"`
	Status         uint8     `db:"status" json:"status"`
	CallbackData   string    `db:"callback_data" json:"callback_data"`
	CreateTime     time.Time `db:"create_time" json:"create_time"`
	UpdateTime     time.Time `db:"update_time" json:"update_time"`
}

// =============== 常量定义 ===============

// 订单状态
const (
	OrderStatusPending uint8 = 0 // 待支付
	OrderStatusPaid    uint8 = 1 // 已支付
	OrderStatusClosed  uint8 = 2 // 已关闭
)

// 支付状态
const (
	PaymentStatusProcessing uint8 = 0 // 处理中
	PaymentStatusSuccess    uint8 = 1 // 成功
	PaymentStatusFailed     uint8 = 2 // 失败
)

// 产品类型
const (
	ProductTypeDay   = "day"   // 日卡
	ProductTypeWeek  = "week"  // 周卡
	ProductTypeMonth = "month" // 月卡
)

// VIP状态
const (
	VipStatusNormal uint8 = 0 // 普通用户
	VipStatusVip    uint8 = 1 // VIP用户
)

// 订阅状态
const (
	SubscriptionStatusActive  uint8 = 1 // 活跃
	SubscriptionStatusExpired uint8 = 0 // 过期
)

// =============== 请求/响应结构体 ===============

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	ProductId     uint32 `json:"product_id" binding:"required"`
	ProductType   string `json:"product_type" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
	ReturnUrl     string `json:"return_url"`
}

// OrderDetailResponse 订单详情响应
type OrderDetailResponse struct {
	Order  *Order `json:"order"`
	PayUrl string `json:"payUrl,omitempty"`
	QrCode string `json:"qrCode,omitempty"`
}

// UserSubscriptionResponse 用户订阅响应
type UserSubscriptionResponse struct {
	VipStatus    uint8      `json:"vip_status"`
	ExpireTime   *time.Time `json:"expire_time"`
	IsActive     bool       `json:"is_active"`
	DaysLeft     int        `json:"days_left"`
	CurrentLevel string     `json:"current_level"`
}

// PaymentCallbackRequest 支付回调请求
type PaymentCallbackRequest struct {
	Pid         string `form:"pid" json:"pid"`
	TradeNo     string `form:"trade_no" json:"trade_no"`
	OutTradeNo  string `form:"out_trade_no" json:"out_trade_no"`
	Type        string `form:"type" json:"type"`
	Name        string `form:"name" json:"name"`
	Money       string `form:"money" json:"money"`
	TradeStatus string `form:"trade_status" json:"trade_status"`
	Param       string `form:"param" json:"param"`
	Sign        string `form:"sign" json:"sign"`
	SignType    string `form:"sign_type" json:"sign_type"`
} 