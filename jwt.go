package main

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var secret string = conf.JwtSecret

type CustomClaims struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.StandardClaims
}

//generate a new token for a user, it last 15 minutre
func NewCustomClaims(username, email string) CustomClaims {
	token := CustomClaims{
		Username: username,
		Email:    email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * time.Duration(15)).Unix(),
			Issuer:    "gitter",
		},
	}
	return token
}

func NewSignedToken(claim CustomClaims) (string, error) {
	//unsigned token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	//sign the token
	return token.SignedString([]byte(secret))
}

func GenerateJWT(username, email string) (string, error) {
	claims := NewCustomClaims(username, email)
	return NewSignedToken(claims)
}

//returns the username and the email from the jwt
func ParseJWT(t string) (string, string, error) {
	token, err := jwt.ParseWithClaims(
		t,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		},
	)
	if err != nil {
		return "", "", err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return "", "", errors.New("couldn't parse claims")
	}

	if claims.ExpiresAt < time.Now().UTC().Unix() {
		return "", "", errors.New("jwt is expired")
	}
	username := claims.Username
	email := claims.Email
	return username, email, nil
}
