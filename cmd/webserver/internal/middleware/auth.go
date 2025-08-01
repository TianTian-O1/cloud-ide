package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/code"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/service"
	"github.com/mangohow/cloud-ide/pkg/logger"
	"github.com/mangohow/cloud-ide/pkg/serialize"
	"github.com/mangohow/cloud-ide/pkg/utils/encrypt"
)

func Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			logger.Logger().Warningf("未获得授权, ip:%s", ctx.Request.RemoteAddr)
			ctx.Status(http.StatusUnauthorized)
			ctx.Abort()
			return
		}

		// 检查并提取Bearer token
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			logger.Logger().Warningf("无效的授权格式, ip:%s", ctx.Request.RemoteAddr)
			ctx.Status(http.StatusUnauthorized)
			ctx.Abort()
			return
		}
		
		token := authHeader[7:] // 去掉"Bearer "前缀

		username, uid, id, err := encrypt.VerifyToken(token)
		if err != nil {
			ctx.Status(http.StatusUnauthorized)
			ctx.Abort()
			return
		}
		ctx.Set("id", id)
		ctx.Set("user_id", id) // 兼容性设置
		ctx.Set("username", username)
		ctx.Set("uid", uid)

		ctx.Next()
	}
}

// VipRequired VIP权限检查中间件
func VipRequired() gin.HandlerFunc {
	subscriptionService := service.NewSubscriptionService()
	
	return func(ctx *gin.Context) {
		// 获取用户ID
		userIdVal, exists := ctx.Get("user_id")
		if !exists {
			ctx.JSON(http.StatusUnauthorized, serialize.Fail(code.LoginFailed))
			ctx.Abort()
			return
		}

		userId, ok := userIdVal.(uint32)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, serialize.Fail(code.LoginFailed))
			ctx.Abort()
			return
		}

		// 检查VIP状态
		isVip := subscriptionService.IsUserVip(userId)
		if !isVip {
			ctx.JSON(http.StatusForbidden, gin.H{"status": code.QueryFailed, "message": "此功能仅限VIP用户使用"})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// CheckWorkspacePermission 工作空间权限检查中间件
func CheckWorkspacePermission(action string) gin.HandlerFunc {
	subscriptionService := service.NewSubscriptionService()
	
	return func(ctx *gin.Context) {
		// 获取用户ID
		userIdVal, exists := ctx.Get("user_id")
		if !exists {
			ctx.JSON(http.StatusUnauthorized, serialize.Fail(code.LoginFailed))
			ctx.Abort()
			return
		}

		userId, ok := userIdVal.(uint32)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, serialize.Fail(code.LoginFailed))
			ctx.Abort()
			return
		}

		// 检查操作权限
		hasPermission := subscriptionService.CheckWorkspacePermission(userId, action)
		if !hasPermission {
			ctx.JSON(http.StatusForbidden, gin.H{"status": code.QueryFailed, "message": "权限不足，请升级为VIP用户"})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}