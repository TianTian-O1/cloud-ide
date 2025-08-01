package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/code"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/model"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/service"
	"github.com/mangohow/cloud-ide/pkg/logger"
	"github.com/mangohow/cloud-ide/pkg/serialize"
	"github.com/sirupsen/logrus"
)

type PaymentController struct {
	logger         *logrus.Logger
	paymentService *service.PaymentService
}

func NewPaymentController() *PaymentController {
	return &PaymentController{
		logger:         logger.Logger(),
		paymentService: service.NewPaymentService(),
	}
}

// GetPaymentProducts 获取支付产品列表
// method: GET path: /api/payment/products
func (p *PaymentController) GetPaymentProducts(ctx *gin.Context) *serialize.Response {
	products, err := p.paymentService.GetPaymentProducts()
	if err != nil {
		p.logger.Errorf("get payment products failed: %v", err)
		return serialize.Fail(code.QueryFailed)
	}

	return serialize.OkData(products)
}

// CreateOrder 创建订单
// method: POST path: /api/payment/order
func (p *PaymentController) CreateOrder(ctx *gin.Context) *serialize.Response {
	var req model.CreateOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		p.logger.Errorf("bind create order request failed: %v", err)
		return serialize.Error(http.StatusBadRequest)
	}

	// 从token中获取用户ID
	userIdVal, exists := ctx.Get("user_id")
	if !exists {
		p.logger.Error("user_id not found in context")
		return serialize.Fail(code.LoginFailed)
	}

	userId, ok := userIdVal.(uint32)
	if !ok {
		p.logger.Error("user_id type assertion failed")
		return serialize.Fail(code.LoginFailed)
	}

	// 设置默认支付方式
	if req.PaymentMethod == "" {
		req.PaymentMethod = "alipay" // 默认支付宝
	}

	// 设置默认跳转地址
	if req.ReturnUrl == "" {
		req.ReturnUrl = "https://tiantianai.co/idea/#/payment"
	}

	orderDetail, err := p.paymentService.CreateOrder(userId, &req)
	if err != nil {
		p.logger.Errorf("create order failed: %v", err)
		switch err {
		case service.ErrProductNotFound:
			return serialize.FailData(code.QueryFailed, gin.H{"message": "产品不存在"})
		case service.ErrPaymentFailed:
			return serialize.FailData(code.QueryFailed, gin.H{"message": "支付创建失败"})
		default:
			return serialize.Fail(code.QueryFailed)
		}
	}

	return serialize.OkData(orderDetail)
}

// GetOrders 获取用户订单列表
// method: GET path: /api/payment/orders
func (p *PaymentController) GetOrders(ctx *gin.Context) *serialize.Response {
	// 从token中获取用户ID
	userIdVal, exists := ctx.Get("user_id")
	if !exists {
		p.logger.Error("user_id not found in context")
		return serialize.Fail(code.LoginFailed)
	}

	userId, ok := userIdVal.(uint32)
	if !ok {
		p.logger.Error("user_id type assertion failed")
		return serialize.Fail(code.LoginFailed)
	}

	// 分页参数
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	orders, err := p.paymentService.GetOrdersByUserId(userId, page, pageSize)
	if err != nil {
		p.logger.Errorf("get orders failed: %v", err)
		return serialize.Fail(code.QueryFailed)
	}

	return serialize.OkData(orders)
}

// GetSubscription 获取用户订阅信息
// method: GET path: /api/payment/subscription
func (p *PaymentController) GetSubscription(ctx *gin.Context) *serialize.Response {
	// 从token中获取用户ID
	userIdVal, exists := ctx.Get("user_id")
	if !exists {
		p.logger.Error("user_id not found in context")
		return serialize.Fail(code.LoginFailed)
	}

	userId, ok := userIdVal.(uint32)
	if !ok {
		p.logger.Error("user_id type assertion failed")
		return serialize.Fail(code.LoginFailed)
	}

	subscription, err := p.paymentService.GetUserSubscription(userId)
	if err != nil {
		p.logger.Errorf("get user subscription failed: %v", err)
		return serialize.Fail(code.QueryFailed)
	}

	return serialize.OkData(subscription)
}

// PaymentCallback 支付回调处理
// method: POST/GET path: /api/payment/callback
func (p *PaymentController) PaymentCallback(ctx *gin.Context) *serialize.Response {
	var req model.PaymentCallbackRequest
	
	// 支持GET、POST form和POST JSON三种方式的回调
	contentType := ctx.GetHeader("Content-Type")
	
	if ctx.Request.Method == "GET" {
		// GET方式，从query参数获取（ouyun支付网关回调使用）
		req.Pid = ctx.Query("pid")
		req.TradeNo = ctx.Query("trade_no")
		req.OutTradeNo = ctx.Query("out_trade_no")
		req.Type = ctx.Query("type")
		req.Name = ctx.Query("name")
		req.Money = ctx.Query("money")
		req.TradeStatus = ctx.Query("trade_status")
		req.Param = ctx.Query("param")
		req.Sign = ctx.Query("sign")
		req.SignType = ctx.Query("sign_type")
		
		p.logger.Infof("GET callback from payment gateway: %+v", req)
		
	} else if strings.Contains(contentType, "application/json") {
		// POST JSON方式，从JSON body获取（前端验证使用）
		if err := ctx.ShouldBindJSON(&req); err != nil {
			p.logger.Errorf("bind payment callback JSON request failed: %v", err)
			return serialize.Error(http.StatusBadRequest)
		}
		
		p.logger.Infof("POST JSON callback from frontend: %+v", req)
		
	} else {
		// POST form方式，从form参数获取（ouyun支付网关回调使用）
		if err := ctx.ShouldBind(&req); err != nil {
			p.logger.Errorf("bind payment callback form request failed: %v", err)
			ctx.String(http.StatusBadRequest, "fail")
			return nil
		}
		
		p.logger.Infof("POST form callback from payment gateway: %+v", req)
	}

	p.logger.Infof("processing payment callback: trade_no=%s, out_trade_no=%s, status=%s", 
		req.TradeNo, req.OutTradeNo, req.TradeStatus)

	err := p.paymentService.HandlePaymentCallback(&req)
	if err != nil {
		p.logger.Errorf("handle payment callback failed: %v", err)
		
		// 如果是JSON请求（前端），返回JSON响应
		if strings.Contains(contentType, "application/json") {
			return serialize.FailData(code.PaymentFailed, gin.H{"message": err.Error()})
		}
		
		// 如果是支付网关回调，返回fail字符串
		ctx.String(http.StatusOK, "fail")
		return nil
	}

	// 如果是JSON请求（前端），返回JSON响应
	if strings.Contains(contentType, "application/json") {
		return serialize.Ok()
	}
	
	// 支付网关回调必须返回"success"表示处理成功
	ctx.String(http.StatusOK, "success")
	return nil
}

// PaymentReturn 支付成功跳转处理
// method: GET path: /api/payment/return
func (p *PaymentController) PaymentReturn(ctx *gin.Context) *serialize.Response {
	orderNo := ctx.Query("out_trade_no")
	tradeNo := ctx.Query("trade_no")
	
	p.logger.Infof("payment return: order_no=%s, trade_no=%s", orderNo, tradeNo)
	
	// 这里可以显示支付成功页面或重定向
	ctx.Redirect(http.StatusFound, "https://tiantianai.co/idea/static/payment-success.html?order_no="+orderNo)
	return nil
}

// SyncPaymentStatus 手动同步支付状态
// method: POST path: /api/payment/sync
func (p *PaymentController) SyncPaymentStatus(ctx *gin.Context) *serialize.Response {
	// 从token中获取用户ID
	userIdVal, exists := ctx.Get("user_id")
	if !exists {
		p.logger.Error("user_id not found in context")
		return serialize.Fail(code.LoginFailed)
	}

	userId, ok := userIdVal.(uint32)
	if !ok {
		p.logger.Error("user_id type assertion failed")
		return serialize.Fail(code.LoginFailed)
	}

	var req struct {
		OrderNo string `json:"order_no" binding:"required"`
	}
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		p.logger.Errorf("bind sync payment request failed: %v", err)
		return serialize.Error(http.StatusBadRequest)
	}

	p.logger.Infof("user %d requesting payment sync for order: %s", userId, req.OrderNo)

	// 模拟支付成功回调
	callbackReq := &model.PaymentCallbackRequest{
		Pid:         "47",
		TradeNo:     fmt.Sprintf("SYNC_%d", time.Now().Unix()),
		OutTradeNo:  req.OrderNo,
		Type:        "alipay",
		Name:        "VIP支付",
		Money:       "9.90", // 这里可以从订单中获取实际金额
		TradeStatus: "TRADE_SUCCESS",
		Param:       "",
		Sign:        "manual_sync",
		SignType:    "MD5",
	}

	err := p.paymentService.HandlePaymentCallback(callbackReq)
	if err != nil {
		p.logger.Errorf("sync payment status failed: %v", err)
		return serialize.FailData(code.PaymentFailed, gin.H{"message": err.Error()})
	}

	p.logger.Infof("payment status synced successfully for order: %s", req.OrderNo)
	return serialize.OkData(gin.H{"message": "支付状态同步成功"})
} 