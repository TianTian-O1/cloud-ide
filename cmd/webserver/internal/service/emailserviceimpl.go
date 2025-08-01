package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/smtp"
	"time"

	"github.com/jordan-wright/email"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/conf"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/dao"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/dao/rdis"
	"github.com/mangohow/cloud-ide/pkg/logger"
	"github.com/mangohow/cloud-ide/pkg/utils"
	"github.com/sirupsen/logrus"
)

type EmailConfig struct {
	host   string
	port   uint32
	sender string
	auth   string
}

type EmailServiceImpl struct {
	logger *logrus.Logger
	ch     chan *email.Email
	pool   *email.Pool
	config *EmailConfig
	dao    *dao.UserDao
}

func NewEmailService() EmailService {
	// 初始化redis
	if err := rdis.InitRedis(); err != nil {
		panic(fmt.Errorf("init redis failed, reason:%s", err.Error()))
	}

	return &EmailServiceImpl{
		logger: logger.Logger(),
		ch:     make(chan *email.Email, 1024),
		config: &EmailConfig{
			host:   conf.EmailConfig.Host,
			port:   conf.EmailConfig.Port,
			sender: conf.EmailConfig.SenderEmail,
			auth:   conf.EmailConfig.AuthCode,
		},
		dao: dao.NewUserDao(),
	}
}

var numbers = []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

func (e *EmailServiceImpl) Send(addr string) error {
	// 生成6位数验证码
	validateCode := make([]byte, 6)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 6; i++ {
		validateCode[i] = numbers[rand.Intn(10)]
	}

	e.logger.Debugf("adrr:%s, code:%s", addr, string(validateCode))
	m := &email.Email{
		From:    e.config.sender,
		To:      []string{addr},
		Subject: "Cloud Code验证码",
		Text:    append([]byte("您好,您的验证码为:"), validateCode...),
		Sender:  "Cloud Code",
	}

	// 存入redis
	if err := rdis.RedisInstance().Set(context.Background(), addr, string(validateCode), time.Minute*5).Err(); err != nil {
		e.logger.Errorf("add validate code err=%v", err)
		return err
	}

	// 发送邮件
	e.ch <- m

	return nil
}

func (e *EmailServiceImpl) Start() error {
	pool, err := email.NewPool(fmt.Sprintf("%s:%d", e.config.host, e.config.port),
		4, smtp.PlainAuth("", e.config.sender, e.config.auth, e.config.host))
	if err != nil {
		e.logger.Errorf("connect to mail failed, err:%v", err)
		return err
	}
	e.pool = pool

	for i := 0; i < 4; i++ {
		go func() {
			for m := range e.ch {
				err := pool.Send(m, 10*time.Second)
				if err != nil {
					e.logger.Errorf("send email error:%v", err)
				}
			}

		}()
	}

	return nil
}

var (
	ErrVerifyFailed = errors.New("验证失败")
	ErrEmailInvalid = errors.New("邮箱不合法")
	ErrCodeInvalid  = errors.New("验证码不合法")
)

func (e *EmailServiceImpl) VerifyEmailValidateCode(email string, code string) error {
	// 验证邮箱有效性
	if !utils.VerifyEmailFormat(email) {
		return ErrEmailInvalid
	}

	// 验证EmailCode长度
	if len(code) != 6 {
		return ErrCodeInvalid
	}

	cmd := rdis.RedisInstance().Get(context.Background(), email)
	if err := cmd.Err(); err != nil {
		return err
	}
	if code != cmd.Val() {
		return ErrVerifyFailed
	}

	return nil
}

func (e *EmailServiceImpl) IsEmailAvailable(email string) bool {
	err := e.dao.FindByEmail(email)
	if err == nil { // 如果err == nil说明查找到了记录
		return false
	}
	return true
}
