package service

import (
	"errors"
	"time"

	"github.com/mangohow/cloud-ide/cmd/webserver/internal/code"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/dao"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/model"
	"github.com/mangohow/cloud-ide/pkg/logger"
	"github.com/mangohow/cloud-ide/pkg/utils/encrypt"
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
)

type UserService struct {
	logger       *logrus.Logger
	dao          *dao.UserDao
	emailService EmailService
}

func NewUserService(service EmailService) *UserService {
	return &UserService{
		logger:       logger.Logger(),
		dao:          dao.NewUserDao(),
		emailService: service,
	}
}

var (
	ErrUserDeleted       = errors.New("user deleted")
	ErrUserNotExist      = errors.New("user not exist")
	ErrPasswordIncorrect = errors.New("password incorrect")
)

func (u *UserService) Login(username, password string) (*model.User, error) {
	u.logger.Infof("UserService.Login: Starting login for user: %s", username)
	
	// 1、从数据库中查询
	u.logger.Infof("UserService.Login: About to call dao.FindByUsernameDetailed for user: %s", username)
	user, err := u.dao.FindByUsernameDetailed(username)
	u.logger.Infof("UserService.Login: dao.FindByUsernameDetailed returned for user: %s, error: %v", username, err)
	if err != nil {
		u.logger.Warnf("UserService.Login: FindByUsernameDetailed failed for user: %s, error: %v", username, err)
		return nil, ErrUserNotExist
	}

	u.logger.Infof("UserService.Login: Found user: %s, starting password verification", username)
	
	// 2、验证密码（兼容明文和加密密码）
	var ok bool
	if len(user.Password) < 32 {
		// 明文密码兼容（旧格式）
		ok = (password == user.Password)
		u.logger.Infof("Using plaintext password verification for user: %s", username)
	} else {
		// 加密密码验证（新格式）
		defer func() {
			if r := recover(); r != nil {
				u.logger.Warnf("Password verification panic, falling back to plaintext: %v", r)
				ok = (password == user.Password)
			}
		}()
		ok = encrypt.VerifyPasswd(password, user.Password)
	}
	
	u.logger.Infof("UserService.Login: Password verification result for user: %s, ok: %v", username, ok)
	if !ok {
		return nil, ErrPasswordIncorrect
	}

	// 3、检查用户状态是否正常
	u.logger.Infof("UserService.Login: Checking user status for user: %s, status: %v", username, user.Status)
	if code.UserStatus(user.Status) == code.StatusDeleted {
		return nil, ErrUserDeleted
	}

	// 4、生成token
	u.logger.Infof("UserService.Login: About to generate token for user: %s", username)
	token, err := encrypt.CreateToken(user.Id, user.Username, user.Uid)
	if err != nil {
		u.logger.Warnf("UserService.Login: Token generation failed for user: %s, error: %v", username, err)
		return nil, err
	}
	user.Token = token

	u.logger.Infof("UserService.Login: Login successful for user: %s", username)
	return user, nil
}

func (u *UserService) CheckUsernameAvailable(username string) bool {
	err := u.dao.FindByUsername(username)
	// 如果能查询到记录， err == nil
	if err != nil {
		return false
	}

	return true
}

var (
	ErrEmailCodeIncorrect = errors.New("email code incorrect")
	ErrEmailAlreadyInUse  = errors.New("this email had been registered")
)

func (u *UserService) UserRegister(info *model.RegisterInfo) error {
	// 1.验证邮箱验证码
	err := u.emailService.VerifyEmailValidateCode(info.Email, info.EmailCode)
	if err != nil {
		u.logger.Infof("verify email code failed err:%v", err)
		return ErrEmailCodeIncorrect
	}

	// 2.验证是否已经存在账号与该邮箱关联，一个邮箱只能创建一个账号
	if !u.emailService.IsEmailAvailable(info.Email) { // 如果err == nil说明查找到了记录
		return ErrEmailAlreadyInUse
	}

	encryptedPasswd := encrypt.PasswdEncrypt(info.Password)

	// 3.生成新的用户
	now := time.Now()
	// 设置删除时间为遥远的未来，表示用户未被删除
	deleteTime := time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)
	user := &model.User{
		Uid:        bson.NewObjectId().Hex(),
		Username:   info.Username,
		Password:   encryptedPasswd,
		Nickname:   info.Nickname,
		Email:      info.Email,
		CreateTime: now,
		DeleteTime: deleteTime,
	}

	err = u.dao.AddUser(user)
	if err != nil {
		return err
	}

	u.logger.Infof("user registered successfully: %s", info.Username)
	return nil
}

// CheckEmailExists 检查邮箱是否存在
func (u *UserService) CheckEmailExists(email string) bool {
	err := u.dao.FindByEmail(email)
	// 如果能查询到记录，err == nil，说明邮箱存在
	return err == nil
}

// ResetPassword 重置密码
func (u *UserService) ResetPassword(email, emailCode, newPassword string) error {
	// 1. 验证邮箱验证码
	err := u.emailService.VerifyEmailValidateCode(email, emailCode)
	if err != nil {
		u.logger.Infof("verify email code failed for reset password err:%v", err)
		return ErrEmailCodeIncorrect
	}

	// 2. 检查用户是否存在
	if !u.CheckEmailExists(email) {
		return ErrUserNotExist
	}

	// 3. 加密新密码
	encryptedPasswd := encrypt.PasswdEncrypt(newPassword)

	// 4. 更新数据库中的密码
	err = u.dao.UpdatePasswordByEmail(email, encryptedPasswd)
	if err != nil {
		u.logger.Errorf("update password failed: %v", err)
		return err
	}

	u.logger.Infof("password reset successfully for email: %s", email)
	return nil
}
