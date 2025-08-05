package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/code"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/service"
	"github.com/mangohow/cloud-ide/pkg/logger"
	"github.com/mangohow/cloud-ide/pkg/serialize"
	"github.com/mangohow/cloud-ide/pkg/utils/encrypt"
	"github.com/sirupsen/logrus"
)

type InternalController struct {
	logger       *logrus.Logger
	spaceService *service.CloudCodeService
}

func NewInternalController() *InternalController {
	return &InternalController{
		logger:       logger.Logger(),
		spaceService: service.NewCloudCodeService(),
	}
}

// VerifyWorkspaceAccessRequest 验证工作空间访问请求
type VerifyWorkspaceAccessRequest struct {
	Sid string `json:"sid" binding:"required"`
}

// VerifyWorkspaceAccess 验证用户是否有权限访问指定的工作空间
// method: POST path: /api/internal/verify-workspace-access
// 这是内部API，仅供网关调用
func (c *InternalController) VerifyWorkspaceAccess(ctx *gin.Context) *serialize.Response {
	// 1. 验证JWT Token
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		c.logger.Warnf("Missing Authorization header from IP: %s", ctx.Request.RemoteAddr)
		ctx.Status(http.StatusUnauthorized)
		return serialize.Fail(code.LoginFailed)
	}

	// 检查并提取Bearer token
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		c.logger.Warnf("Invalid Authorization format from IP: %s", ctx.Request.RemoteAddr)
		ctx.Status(http.StatusUnauthorized)
		return serialize.Fail(code.LoginFailed)
	}
	
	token := authHeader[7:] // 去掉"Bearer "前缀

	_, uid, _, err := encrypt.VerifyToken(token)
	if err != nil {
		c.logger.Warnf("Invalid token from IP: %s, error: %v", ctx.Request.RemoteAddr, err)
		ctx.Status(http.StatusUnauthorized)
		return serialize.Fail(code.LoginFailed)
	}

	// 2. 解析请求参数
	var req VerifyWorkspaceAccessRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Errorf("Invalid request parameters: %v", err)
		ctx.Status(http.StatusBadRequest)
		return serialize.Fail(code.QueryFailed)
	}

	// 3. 验证用户是否拥有该工作空间
	hasAccess, err := c.spaceService.VerifyWorkspaceAccess(uid, req.Sid)
	if err != nil {
		c.logger.Errorf("Failed to verify workspace access for user %s, sid %s: %v", uid, req.Sid, err)
		ctx.Status(http.StatusInternalServerError)
		return serialize.Fail(code.QueryFailed)
	}

	if !hasAccess {
		c.logger.Warnf("Access denied: user %s attempted to access workspace %s", uid, req.Sid)
		ctx.Status(http.StatusForbidden)
		return serialize.Fail(code.PermissionDenied)
	}

	c.logger.Infof("Access granted: user %s accessing workspace %s", uid, req.Sid)
	return serialize.Ok()
}