package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mangohow/cloud-ide/cmd/webserver/internal/conf"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/dao"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/model"
	"github.com/mangohow/cloud-ide/pkg/logger"
	"github.com/mangohow/cloud-ide/pkg/utils/encrypt"
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
)

type OAuthService struct {
	logger  *logrus.Logger
	userDao *dao.UserDao
}

func NewOAuthService() *OAuthService {
	return &OAuthService{
		logger:  logger.Logger(),
		userDao: dao.NewUserDao(),
	}
}

// LinuxDoUserInfo LinuxDo用户信息结构
type LinuxDoUserInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar_url"`
}

// LinuxDoTokenResponse LinuxDo token响应结构
type LinuxDoTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

var (
	ErrOAuthTokenExchange = errors.New("oauth token exchange failed")
	ErrOAuthUserInfo     = errors.New("oauth user info retrieval failed")
	ErrOAuthUserCreate   = errors.New("oauth user creation failed")
)

// GetLinuxDoAuthURL 获取LinuxDo OAuth授权URL
func (o *OAuthService) GetLinuxDoAuthURL(state string) string {
	baseURL := conf.OAuthConfig.LinuxDoBaseURL
	clientID := conf.OAuthConfig.LinuxDoClientID
	redirectURL := conf.OAuthConfig.LinuxDoRedirectURL
	
	params := url.Values{}
	params.Add("client_id", clientID)
	params.Add("redirect_uri", redirectURL)
	params.Add("response_type", "code")
	params.Add("scope", "read")
	params.Add("state", state)
	
	return fmt.Sprintf("%s/oauth2/authorize?%s", baseURL, params.Encode())
}

// ExchangeLinuxDoToken 使用授权码交换访问令牌
func (o *OAuthService) ExchangeLinuxDoToken(code string) (*LinuxDoTokenResponse, error) {
	baseURL := conf.OAuthConfig.LinuxDoBaseURL
	clientID := conf.OAuthConfig.LinuxDoClientID
	clientSecret := conf.OAuthConfig.LinuxDoClientSecret
	redirectURL := conf.OAuthConfig.LinuxDoRedirectURL
	
	// 准备请求参数
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", redirectURL)
	
	// 发送POST请求
	resp, err := http.PostForm(fmt.Sprintf("%s/oauth2/token", baseURL), data)
	if err != nil {
		o.logger.Errorf("Failed to exchange token: %v", err)
		return nil, ErrOAuthTokenExchange
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		o.logger.Errorf("Token exchange failed with status: %d", resp.StatusCode)
		return nil, ErrOAuthTokenExchange
	}
	
	// 解析响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		o.logger.Errorf("Failed to read token response: %v", err)
		return nil, ErrOAuthTokenExchange
	}
	
	var tokenResp LinuxDoTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		o.logger.Errorf("Failed to parse token response: %v", err)
		return nil, ErrOAuthTokenExchange
	}
	
	return &tokenResp, nil
}

// GetLinuxDoUserInfo 获取LinuxDo用户信息
func (o *OAuthService) GetLinuxDoUserInfo(accessToken string) (*LinuxDoUserInfo, error) {
	baseURL := conf.OAuthConfig.LinuxDoBaseURL
	
	// 创建请求
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/user", baseURL), nil)
	if err != nil {
		o.logger.Errorf("Failed to create user info request: %v", err)
		return nil, ErrOAuthUserInfo
	}
	
	// 添加授权头
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")
	
	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		o.logger.Errorf("Failed to get user info: %v", err)
		return nil, ErrOAuthUserInfo
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		o.logger.Errorf("User info request failed with status: %d", resp.StatusCode)
		return nil, ErrOAuthUserInfo
	}
	
	// 解析响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		o.logger.Errorf("Failed to read user info response: %v", err)
		return nil, ErrOAuthUserInfo
	}
	
	var userInfo LinuxDoUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		o.logger.Errorf("Failed to parse user info response: %v", err)
		return nil, ErrOAuthUserInfo
	}
	
	// 调试：记录LinuxDo返回的用户信息
	o.logger.Infof("LinuxDo user info: ID=%d, Username=%s, Name=%s, Email=%s", 
		userInfo.ID, userInfo.Username, userInfo.Name, userInfo.Email)
	
	return &userInfo, nil
}

// LoginOrCreateUser 通过OAuth信息登录或创建用户
func (o *OAuthService) LoginOrCreateUser(userInfo *LinuxDoUserInfo) (*model.User, error) {
	// 首先尝试通过LinuxDo ID查找用户
	user, err := o.userDao.FindByLinuxDoID(userInfo.ID)
	if err == nil {
		// 用户已存在，更新信息并返回
		return o.updateAndLoginUser(user, userInfo)
	}
	
	// 尝试通过邮箱查找用户
	if userInfo.Email != "" {
		user, err = o.userDao.FindByEmailDetailed(userInfo.Email)
		if err == nil {
			// 用户存在但没有绑定LinuxDo，绑定并返回
			return o.bindLinuxDoAndLogin(user, userInfo)
		}
	}
	
	// 用户不存在，创建新用户
	return o.createNewOAuthUser(userInfo)
}

// updateAndLoginUser 更新用户信息并登录
func (o *OAuthService) updateAndLoginUser(user *model.User, userInfo *LinuxDoUserInfo) (*model.User, error) {
	// 更新用户信息
	updates := map[string]interface{}{
		"nickname":       userInfo.Username, // 使用LinuxDo用户名作为昵称
		"avatar":         userInfo.Avatar,
		"linuxdo_username": userInfo.Username,
	}
	
	if userInfo.Email != "" && user.Email != userInfo.Email {
		updates["email"] = userInfo.Email
	}
	
	err := o.userDao.UpdateUser(user.Id, updates)
	if err != nil {
		o.logger.Errorf("Failed to update user info: %v", err)
	}
	
	// 生成token
	token, err := encrypt.CreateToken(user.Id, user.Username, user.Uid)
	if err != nil {
		o.logger.Errorf("Failed to create token: %v", err)
		return nil, err
	}
	
	user.Token = token
	user.Nickname = userInfo.Username // 使用LinuxDo用户名作为昵称
	user.Avatar = userInfo.Avatar
	
	return user, nil
}

// bindLinuxDoAndLogin 绑定LinuxDo账号并登录
func (o *OAuthService) bindLinuxDoAndLogin(user *model.User, userInfo *LinuxDoUserInfo) (*model.User, error) {
	// 绑定LinuxDo信息
	updates := map[string]interface{}{
		"linuxdo_id":       userInfo.ID,
		"linuxdo_username": userInfo.Username,
		"nickname":         userInfo.Username, // 使用LinuxDo用户名作为昵称
		"avatar":           userInfo.Avatar,
	}
	
	err := o.userDao.UpdateUser(user.Id, updates)
	if err != nil {
		o.logger.Errorf("Failed to bind LinuxDo account: %v", err)
		return nil, ErrOAuthUserCreate
	}
	
	// 生成token
	token, err := encrypt.CreateToken(user.Id, user.Username, user.Uid)
	if err != nil {
		o.logger.Errorf("Failed to create token: %v", err)
		return nil, err
	}
	
	user.Token = token
	user.Nickname = userInfo.Username // 使用LinuxDo用户名作为昵称
	user.Avatar = userInfo.Avatar
	
	return user, nil
}

// createNewOAuthUser 创建新的OAuth用户
func (o *OAuthService) createNewOAuthUser(userInfo *LinuxDoUserInfo) (*model.User, error) {
	now := time.Now()
	deleteTime := time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)
	
	// 生成唯一的用户名，如果LinuxDo用户名已存在则添加后缀
	username := o.generateUniqueUsername(userInfo.Username)
	
	user := &model.User{
		Uid:              bson.NewObjectId().Hex(),
		Username:         username,
		Password:         "", // OAuth用户不需要密码
		Nickname:         userInfo.Username, // 使用LinuxDo用户名作为昵称
		Email:            userInfo.Email,
		Avatar:           userInfo.Avatar,
		CreateTime:       now,
		DeleteTime:       deleteTime,
		Status:           0, // 正常状态
		LinuxDoID:        &userInfo.ID,
		LinuxDoUsername:  &userInfo.Username,
	}
	
	err := o.userDao.AddUser(user)
	if err != nil {
		o.logger.Errorf("Failed to create OAuth user: %v", err)
		return nil, ErrOAuthUserCreate
	}
	
	// 生成token
	token, err := encrypt.CreateToken(user.Id, user.Username, user.Uid)
	if err != nil {
		o.logger.Errorf("Failed to create token: %v", err)
		return nil, err
	}
	
	user.Token = token
	
	o.logger.Infof("Created new OAuth user: %s (LinuxDo: %s)", username, userInfo.Username)
	return user, nil
}

// generateUniqueUsername 生成唯一的用户名
func (o *OAuthService) generateUniqueUsername(baseUsername string) string {
	// 清理用户名，只保留字母数字和下划线
	username := strings.ToLower(baseUsername)
	username = strings.ReplaceAll(username, " ", "_")
	username = strings.ReplaceAll(username, "-", "_")
	
	// 检查用户名是否可用
	if o.isUsernameAvailable(username) {
		return username
	}
	
	// 如果不可用，添加数字后缀
	for i := 1; i <= 999; i++ {
		newUsername := fmt.Sprintf("%s_%d", username, i)
		if o.isUsernameAvailable(newUsername) {
			return newUsername
		}
	}
	
	// 如果还是不可用，使用时间戳
	return fmt.Sprintf("%s_%d", username, time.Now().Unix())
}

// isUsernameAvailable 检查用户名是否可用
func (o *OAuthService) isUsernameAvailable(username string) bool {
	err := o.userDao.FindByUsername(username)
	return err != nil // 如果查询出错，说明用户名不存在，可用
}