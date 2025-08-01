package encrypt

import "testing"

func TestSalt(t *testing.T) {
	password := "123456"
	str := password + Salt()
	t.Log(len(str), str)
	md5String := Md5String(str)
	t.Log(len(md5String), md5String)
}

func TestEncryptPasswd(t *testing.T) {
	passwd := "abcdefg"
	encrypted := PasswdEncrypt(passwd)
	t.Log(len(encrypted), encrypted)
	t.Log(VerifyPasswd(passwd, encrypted))
}

func TestEncryptPasswd1(t *testing.T) {
	passwd := "abcdefg"
	encrypted := PasswdEncrypt(passwd)
	t.Log(len(encrypted), encrypted)
	t.Log(VerifyPasswd(passwd+"a", encrypted))
}
