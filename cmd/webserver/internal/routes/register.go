package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/controller"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/middleware"
	"github.com/mangohow/cloud-ide/pkg/router"
)

func Register(engine *gin.Engine) {
	// 添加CORS中间件
	engine.Use(corsMiddleware())
	
	// 添加静态文件服务
	engine.Static("/images", "./static/images")
	engine.Static("/static", "./static")
	
	authGroup := engine.Group("/auth")
	userController := controller.NewUserController()
	oauthController := controller.NewOAuthController()
	{
		authGroup.POST("/login", router.HandlerAdapter(userController.Login))
		authGroup.GET("/username/check", router.HandlerAdapter(userController.CheckUsernameAvailable))
		authGroup.POST("/register", router.HandlerAdapter(userController.Register))
		authGroup.GET("/emailCode", router.HandlerAdapter(userController.GetEmailValidateCode))
		authGroup.POST("/forgot-password", router.HandlerAdapter(userController.ForgotPassword))
		authGroup.POST("/reset-password", router.HandlerAdapter(userController.ResetPassword))
		
		// OAuth相关路由
		authGroup.GET("/oauth/status", router.HandlerAdapter(oauthController.GetOAuthStatus))
		authGroup.GET("/oauth/linuxdo/login", router.HandlerAdapter(oauthController.LinuxDoLogin))
		authGroup.GET("/oauth/linuxdo/callback", router.HandlerAdapter(oauthController.LinuxDoCallback))
		authGroup.POST("/oauth/linuxdo/callback/api", router.HandlerAdapter(oauthController.LinuxDoCallbackAPI))
	}

	apiGroup := engine.Group("/api", middleware.Auth())
	tmplController := controller.NewSpaceTmplController()
	{
		apiGroup.GET("/template/list", router.HandlerAdapter(tmplController.SpaceTmpls))
		apiGroup.GET("/spec/list", router.HandlerAdapter(tmplController.SpaceSpecs))
	}

	spaceController := controller.NewCloudCodeController()
	{
		apiGroup.GET("/workspace/list", router.HandlerAdapter(spaceController.ListSpace))
		apiGroup.DELETE("/workspace", router.HandlerAdapter(spaceController.DeleteSpace))
		apiGroup.POST("/workspace", router.HandlerAdapter(spaceController.CreateSpace))
		apiGroup.POST("/workspace/cas", router.HandlerAdapter(spaceController.CreateSpaceAndStart))
		apiGroup.PUT("/workspace/start", router.HandlerAdapter(spaceController.StartSpace))
		apiGroup.PUT("/workspace/stop", router.HandlerAdapter(spaceController.StopSpace))
		apiGroup.PUT("/workspace/name", router.HandlerAdapter(spaceController.ModifySpaceName))
	}

	// 支付相关路由
	paymentController := controller.NewPaymentController()
	paymentGroup := apiGroup.Group("/payment")
	{
		// 需要认证的支付API
		paymentGroup.POST("/order", router.HandlerAdapter(paymentController.CreateOrder))
		paymentGroup.GET("/orders", router.HandlerAdapter(paymentController.GetOrders))
		paymentGroup.GET("/subscription", router.HandlerAdapter(paymentController.GetSubscription))
		paymentGroup.POST("/sync", router.HandlerAdapter(paymentController.SyncPaymentStatus))
	}

	// 支付回调路由和公开API（不需要认证）
	callbackGroup := engine.Group("/api/payment")
	{
		// 公开API - 不需要认证
		callbackGroup.GET("/products", router.HandlerAdapter(paymentController.GetPaymentProducts))
		// 回调API - 不需要认证
		callbackGroup.Any("/callback", router.HandlerAdapter(paymentController.PaymentCallback))
		callbackGroup.GET("/return", router.HandlerAdapter(paymentController.PaymentReturn))
	}
}

// corsMiddleware 添加CORS支持
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Header("Content-Type", "application/json; charset=utf-8")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}