package service

type EmailService interface {
	Send(addr string) error

	Start() error

	VerifyEmailValidateCode(email string, code string) error

	IsEmailAvailable(email string) bool
}
