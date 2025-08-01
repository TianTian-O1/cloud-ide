package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/code"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/conf"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/model"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/service"
	"github.com/mangohow/cloud-ide/pkg/logger"
	"github.com/mangohow/cloud-ide/pkg/serialize"
	"github.com/mangohow/cloud-ide/pkg/utils"

	"github.com/sirupsen/logrus"
)

type UserController struct {
	logger       *logrus.Logger
	service      *service.UserService
	emailService service.EmailService
}

func NewUserController() *UserController {
	var emailService service.EmailService
	if conf.EmailConfig.Enabled {
		emailService = service.NewEmailService()
	} else {
		emailService = service.NewFakeEmailService()
	}

	err := emailService.Start()
	if err != nil {
		panic(err)
	}
	return &UserController{
		service:      service.NewUserService(emailService),
		logger:       logger.Logger(),
		emailService: emailService,
	}
}

// Login method: POST path: /auth/login
func (u *UserController) Login(ctx *gin.Context) *serialize.Response {
	var userInfo struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := ctx.ShouldBind(&userInfo)
	if err != nil {
		userInfo.Username = ctx.PostForm("username")
		userInfo.Password = ctx.PostForm("password")
	}

	u.logger.Debugf("username:%s passowrd:%s", userInfo.Username, userInfo.Password)
	if userInfo.Username == "" || userInfo.Password == "" {
		return serialize.FailData(code.LoginFailed, nil)
	}
	u.logger.Debugf("username:%s, pasword:%s", userInfo.Username, userInfo.Password)

	user, err := u.service.Login(userInfo.Username, userInfo.Password)
	if err != nil {
		switch err {
		case service.ErrUserDeleted:
			return serialize.Fail(code.LoginUserDeleted)
		case service.ErrUserNotExist:
			return serialize.Fail(code.LoginUserNotExist)
		case service.ErrPasswordIncorrect:
			return serialize.Fail(code.LoginPasswordIncorrect)
		}

		u.logger.Warnf("login error:%v", err)
		return serialize.Fail(code.LoginFailed)
	}

	// 使用正确的LoginSuccess状态码
	return serialize.OkCodeData(code.LoginSuccess, user)
}

// Register 用户注册 method: POST path: /auth/register
func (u *UserController) Register(ctx *gin.Context) *serialize.Response {
	var info model.RegisterInfo
	err := ctx.ShouldBind(&info)
	if err != nil {
		return serialize.Error(http.StatusBadRequest)
	}
	u.logger.Debug("register", info)
	// 验证用户名长度
	if len(info.Username) < 3 || len(info.Username) > 10 {
		return serialize.Fail(code.UserNameLengthInvalid)
	}

	// 验证EmailCode长度
	if len(info.EmailCode) != 6 {
		return serialize.Fail(code.UserEmailCodeInvalid)
	}

	// 验证邮箱有效性
	if !utils.VerifyEmailFormat(info.Email) {
		return serialize.Fail(code.UserEmailInvalid)
	}

	err = u.service.UserRegister(&info)
	switch err {
	case service.ErrEmailCodeIncorrect:
		return serialize.Fail(code.UserEmailCodeIncorrect)
	case service.ErrEmailAlreadyInUse:
		return serialize.Fail(code.UserEmailAlreadyInUse)
	case nil:
		return serialize.Ok()
	}

	u.logger.Debugf("add user err:%v", err)
	return serialize.Fail(code.UserRegisterFailed)
}

// CheckUsernameAvailable 检测用户名是否可用 method: GET path: /auth/username/check
func (u *UserController) CheckUsernameAvailable(ctx *gin.Context) *serialize.Response {
	u.logger.Debugf("check username available")
	value := ctx.Query("username")
	if value == "" {
		return serialize.Error(http.StatusBadRequest)
	}

	ok := u.service.CheckUsernameAvailable(value)
	if !ok {
		return serialize.Ok()
	}

	return serialize.Fail(code.UserNameUnavailable)
}

// GetEmailValidateCode 通过邮箱获取验证码 method: GET path: /auth/emailCode
func (u *UserController) GetEmailValidateCode(ctx *gin.Context) *serialize.Response {
	addr := ctx.Query("email")
	if addr == "" {
		return serialize.Error(http.StatusBadRequest)
	}

	err := u.emailService.Send(addr)
	if err != nil {
		return serialize.Fail(code.UserSendValidateCodeFailed)
	}

	return serialize.Ok()
}

// ForgotPassword 发送忘记密码验证码 method: POST path: /auth/forgot-password
func (u *UserController) ForgotPassword(ctx *gin.Context) *serialize.Response {
	var req struct {
		Email string `json:"email"`
	}
	err := ctx.ShouldBind(&req)
	if err != nil {
		return serialize.Error(http.StatusBadRequest)
	}

	// 验证邮箱格式
	if !utils.VerifyEmailFormat(req.Email) {
		return serialize.Fail(code.UserEmailInvalid)
	}

	// 检查邮箱是否存在
	exists := u.service.CheckEmailExists(req.Email)
	if !exists {
		return serialize.Fail(code.UserEmailNotExists)
	}

	// 发送重置密码验证码
	err = u.emailService.Send(req.Email)
	if err != nil {
		u.logger.Errorf("send forgot password email failed: %v", err)
		return serialize.Fail(code.UserSendValidateCodeFailed)
	}

	return serialize.Ok()
}

// ResetPassword 重置密码 method: POST path: /auth/reset-password
func (u *UserController) ResetPassword(ctx *gin.Context) *serialize.Response {
	var req struct {
		Email       string `json:"email"`
		EmailCode   string `json:"emailCode"`
		NewPassword string `json:"newPassword"`
	}
	err := ctx.ShouldBind(&req)
	if err != nil {
		return serialize.Error(http.StatusBadRequest)
	}

	// 验证邮箱格式
	if !utils.VerifyEmailFormat(req.Email) {
		return serialize.Fail(code.UserEmailInvalid)
	}

	// 验证验证码长度
	if len(req.EmailCode) != 6 {
		return serialize.Fail(code.UserEmailCodeInvalid)
	}

	// 验证新密码长度
	if len(req.NewPassword) < 8 || len(req.NewPassword) > 24 {
		return serialize.Fail(code.UserPasswordLengthInvalid)
	}

	// 重置密码
	err = u.service.ResetPassword(req.Email, req.EmailCode, req.NewPassword)
	switch err {
	case service.ErrEmailCodeIncorrect:
		return serialize.Fail(code.UserEmailCodeIncorrect)
	case service.ErrUserNotExist:
		return serialize.Fail(code.UserEmailNotExists)
	case nil:
		return serialize.Ok()
	}

	u.logger.Errorf("reset password failed: %v", err)
	return serialize.Fail(code.UserResetPasswordFailed)
}
