package controller

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/code"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/service"
	"github.com/mangohow/cloud-ide/pkg/logger"
	"github.com/mangohow/cloud-ide/pkg/serialize"
	"github.com/sirupsen/logrus"
)

type OAuthController struct {
	logger       *logrus.Logger
	oauthService *service.OAuthService
	stateStore   map[string]time.Time // 简单的state存储，生产环境建议使用Redis
}

func NewOAuthController() *OAuthController {
	return &OAuthController{
		logger:       logger.Logger(),
		oauthService: service.NewOAuthService(),
		stateStore:   make(map[string]time.Time),
	}
}

// LinuxDoLogin 发起LinuxDo OAuth登录
// method: GET path: /auth/oauth/linuxdo/login
func (o *OAuthController) LinuxDoLogin(ctx *gin.Context) *serialize.Response {
	// 生成随机state参数防止CSRF攻击
	state, err := o.generateState()
	if err != nil {
		o.logger.Errorf("Failed to generate OAuth state: %v", err)
		return serialize.Error(http.StatusInternalServerError)
	}
	
	// 存储state，设置5分钟有效期
	o.stateStore[state] = time.Now().Add(5 * time.Minute)
	
	// 获取LinuxDo授权URL
	authURL := o.oauthService.GetLinuxDoAuthURL(state)
	
	o.logger.Infof("Redirecting to LinuxDo OAuth: %s", authURL)
	
	// 返回授权URL，让前端处理跳转
	return serialize.OkData(gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// LinuxDoCallback 处理LinuxDo OAuth回调
// method: GET path: /auth/oauth/linuxdo/callback
func (o *OAuthController) LinuxDoCallback(ctx *gin.Context) *serialize.Response {
	code := ctx.Query("code")
	state := ctx.Query("state")
	errorParam := ctx.Query("error")
	
	// 检查是否有错误
	if errorParam != "" {
		o.logger.Warnf("OAuth callback error: %s", errorParam)
		// 重定向到登录页面并显示错误
		ctx.Redirect(http.StatusFound, "https://tiantianai.co/idea/#/login?error=oauth_error")
		return nil
	}
	
	// 验证code和state参数
	if code == "" || state == "" {
		o.logger.Warn("OAuth callback missing code or state parameter")
		ctx.Redirect(http.StatusFound, "https://tiantianai.co/idea/#/login?error=invalid_request")
		return nil
	}
	
	// 验证state参数
	if !o.validateState(state) {
		o.logger.Warn("OAuth callback invalid state parameter")
		ctx.Redirect(http.StatusFound, "https://tiantianai.co/idea/#/login?error=invalid_state")
		return nil
	}
	
	// 交换访问令牌
	tokenResp, err := o.oauthService.ExchangeLinuxDoToken(code)
	if err != nil {
		o.logger.Errorf("Failed to exchange OAuth token: %v", err)
		ctx.Redirect(http.StatusFound, "https://tiantianai.co/idea/#/login?error=token_exchange_failed")
		return nil
	}
	
	// 获取用户信息
	userInfo, err := o.oauthService.GetLinuxDoUserInfo(tokenResp.AccessToken)
	if err != nil {
		o.logger.Errorf("Failed to get OAuth user info: %v", err)
		ctx.Redirect(http.StatusFound, "https://tiantianai.co/idea/#/login?error=user_info_failed")
		return nil
	}
	
	// 登录或创建用户
	user, err := o.oauthService.LoginOrCreateUser(userInfo)
	if err != nil {
		o.logger.Errorf("Failed to login or create OAuth user: %v", err)
		ctx.Redirect(http.StatusFound, "https://tiantianai.co/idea/#/login?error=login_failed")
		return nil
	}
	
	o.logger.Infof("OAuth login successful for user: %s (LinuxDo: %s)", user.Username, userInfo.Username)
	
	// 构建成功跳转URL，包含用户信息
	successURL := fmt.Sprintf("https://tiantianai.co/idea/#/oauth/success?token=%s&username=%s&nickname=%s&user_id=%d", 
		user.Token, user.Username, user.Nickname, user.Id)
	
	ctx.Redirect(http.StatusFound, successURL)
	return nil
}

// LinuxDoCallbackAPI 处理LinuxDo OAuth回调的API版本（用于API调用）
// method: POST path: /auth/oauth/linuxdo/callback/api
func (o *OAuthController) LinuxDoCallbackAPI(ctx *gin.Context) *serialize.Response {
	var req struct {
		Code  string `json:"code"`
		State string `json:"state"`
	}
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return serialize.Error(http.StatusBadRequest)
	}
	
	// 验证code和state参数
	if req.Code == "" || req.State == "" {
		return serialize.FailData(code.LoginFailed, gin.H{"message": "Missing code or state parameter"})
	}
	
	// 验证state参数
	if !o.validateState(req.State) {
		return serialize.FailData(code.LoginFailed, gin.H{"message": "Invalid state parameter"})
	}
	
	// 交换访问令牌
	tokenResp, err := o.oauthService.ExchangeLinuxDoToken(req.Code)
	if err != nil {
		o.logger.Errorf("Failed to exchange OAuth token: %v", err)
		return serialize.FailData(code.LoginFailed, gin.H{"message": "Token exchange failed"})
	}
	
	// 获取用户信息
	userInfo, err := o.oauthService.GetLinuxDoUserInfo(tokenResp.AccessToken)
	if err != nil {
		o.logger.Errorf("Failed to get OAuth user info: %v", err)
		return serialize.FailData(code.LoginFailed, gin.H{"message": "User info retrieval failed"})
	}
	
	// 登录或创建用户
	user, err := o.oauthService.LoginOrCreateUser(userInfo)
	if err != nil {
		o.logger.Errorf("Failed to login or create OAuth user: %v", err)
		return serialize.FailData(code.LoginFailed, gin.H{"message": "Login failed"})
	}
	
	o.logger.Infof("OAuth login successful for user: %s (LinuxDo: %s)", user.Username, userInfo.Username)
	
	return serialize.OkCodeData(code.LoginSuccess, user)
}

// generateState 生成随机state参数
func (o *OAuthController) generateState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// validateState 验证state参数
func (o *OAuthController) validateState(state string) bool {
	// 清理过期的state
	now := time.Now()
	for s, expiry := range o.stateStore {
		if now.After(expiry) {
			delete(o.stateStore, s)
		}
	}
	
	// 检查state是否存在且未过期
	expiry, exists := o.stateStore[state]
	if !exists {
		return false
	}
	
	if now.After(expiry) {
		delete(o.stateStore, state)
		return false
	}
	
	// 使用后删除state
	delete(o.stateStore, state)
	return true
}

// GetOAuthStatus 获取OAuth配置状态
// method: GET path: /auth/oauth/status
func (o *OAuthController) GetOAuthStatus(ctx *gin.Context) *serialize.Response {
	// 检查LinuxDo OAuth是否配置
	linuxdoEnabled := o.oauthService != nil
	
	return serialize.OkData(gin.H{
		"linuxdo_enabled": linuxdoEnabled,
	})
}