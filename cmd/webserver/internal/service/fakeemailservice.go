package service

type FakeEmailService struct {
}

func NewFakeEmailService() EmailService {
	return FakeEmailService{}
}

func (e FakeEmailService) Send(addr string) error {
	return nil
}

func (e FakeEmailService) Start() error {
	return nil
}

func (e FakeEmailService) VerifyEmailValidateCode(email string, code string) error {
	return nil
}

func (e FakeEmailService) IsEmailAvailable(email string) bool {
	return true
}
