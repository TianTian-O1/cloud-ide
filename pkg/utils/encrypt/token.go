package encrypt

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Claim struct {
	Username string
	Id       uint32
	Uid      string
	jwt.StandardClaims
}

var jwtKey = []byte("cloud-ide-webserver")

func CreateToken(id uint32, username, uid string) (string, error) {
	now := time.Now()
	claims := &Claim{
		Username: username,
		Id:       id,
		Uid:      uid,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(time.Hour * 12).Unix(),
			Issuer:    "mgh",
			Subject:   "User_Token",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func VerifyToken(token string) (string, string, uint32, error) {
	if token == "" {
		return "", "", 0, errors.New("empty String")
	}
	data, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return "", "", 0, err
	}

	claim, ok := data.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", 0, errors.New("parse Error")
	}

	// 安全地提取Username
	username, ok := claim["Username"].(string)
	if !ok {
		return "", "", 0, errors.New("invalid Username in token")
	}

	// 安全地提取Uid
	uid, ok := claim["Uid"].(string)
	if !ok {
		return "", "", 0, errors.New("invalid Uid in token")
	}

	// 安全地提取Id，支持多种数字类型
	var id uint32
	switch v := claim["Id"].(type) {
	case float64:
		if v < 0 || v > float64(^uint32(0)) {
			return "", "", 0, errors.New("invalid Id range in token")
		}
		id = uint32(v)
	case float32:
		if v < 0 || v > float32(^uint32(0)) {
			return "", "", 0, errors.New("invalid Id range in token")
		}
		id = uint32(v)
	case int:
		if v < 0 || v > int(^uint32(0)) {
			return "", "", 0, errors.New("invalid Id range in token")
		}
		id = uint32(v)
	case int32:
		if v < 0 {
			return "", "", 0, errors.New("invalid Id range in token")
		}
		id = uint32(v)
	case int64:
		if v < 0 || v > int64(^uint32(0)) {
			return "", "", 0, errors.New("invalid Id range in token")
		}
		id = uint32(v)
	case uint32:
		id = v
	case uint64:
		if v > uint64(^uint32(0)) {
			return "", "", 0, errors.New("invalid Id range in token")
		}
		id = uint32(v)
	default:
		return "", "", 0, errors.New("invalid Id type in token")
	}

	// 验证Id不能为0
	if id == 0 {
		return "", "", 0, errors.New("invalid user Id: cannot be zero")
	}

	return username, uid, id, nil
}
