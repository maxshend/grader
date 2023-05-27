package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type jwtClaims struct {
	jwt.RegisteredClaims
}

const expMinutes = 5

var ErrInvalidAccessToken = errors.New("invalid access token")

func AccessToken(secret, id string) (string, error) {
	data := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        id,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expMinutes * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return signed, nil
}

func CheckAccessToken(secret, tokenString, id string) error {
	payload := &jwtClaims{}
	token, err := jwt.ParseWithClaims(tokenString, payload, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return err
	}

	if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid && claims.ID == id {
		return nil
	}

	return ErrInvalidAccessToken
}
