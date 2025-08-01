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

	return claim["Username"].(string), claim["Uid"].(string), uint32(claim["Id"].(float64)), nil
}
