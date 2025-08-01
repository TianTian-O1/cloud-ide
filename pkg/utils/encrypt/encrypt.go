package encrypt

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/mangohow/cloud-ide/pkg/utils"
)

func Md5String(str string) string {
	return fmt.Sprintf("%x", md5.Sum(utils.String2Bytes(str)))
}

var (
	encryptedPasswordLen = len(PasswdEncrypt("123456"))
)

// PasswdEncrypt 对密码进行加盐加密
func PasswdEncrypt(passwd string) string {
	salt := Salt()
	ep := passwd + salt
	return Md5String(ep) + salt
}

// VerifyPasswd 验证密码
func VerifyPasswd(passwd, encrypted string) bool {
	if len(encrypted) != encryptedPasswordLen {
		panic("encrypted password invalid")
	}
	psd := encrypted[:32]
	salt := encrypted[32:]
	ep := passwd + salt

	return psd == Md5String(ep)
}

// Salt 生成随机盐值
func Salt() string {
	// 定义一个字节数组用于存储生成的随机数
	randomBytes := make([]byte, 10)

	// 使用crypto/rand包生成随机数
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}

	// 将随机数转换为base64编码的字符串
	return base64.StdEncoding.EncodeToString(randomBytes)
}
