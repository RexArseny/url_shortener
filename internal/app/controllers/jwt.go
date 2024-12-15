package controllers

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	authorization = "Authorization"
	maxAge        = 900
)

var ErrNoJWT = errors.New("no jwt in cookie")

type JWT struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
}

func (c *Controller) getJWT(ctx *gin.Context) (*JWT, error) {
	tokenString, err := ctx.Cookie(authorization)
	if err != nil || tokenString == "" {
		userID := uuid.New()
		return &JWT{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "url_shortener",
				Subject:   userID.String(),
				Audience:  jwt.ClaimStrings{},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * maxAge)),
				NotBefore: jwt.NewNumericDate(time.Now()),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ID:        uuid.New().String(),
			},
			UserID: userID,
		}, ErrNoJWT
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWT{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodEdDSA {
				return nil, errors.New("jwt signature mismatch")
			}
			return c.publicKey, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("can not parse jwt: %w", err)
	}

	claims, ok := token.Claims.(*JWT)
	if !ok {
		return nil, errors.New("token is not jwt format")
	}

	return claims, nil
}

func (c *Controller) setJWT(ctx *gin.Context, claims *JWT) error {
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)

	tokenString, err := token.SignedString(c.privateKey)
	if err != nil {
		return fmt.Errorf("can not sign token: %w", err)
	}

	ctx.SetCookie(
		authorization,
		tokenString,
		maxAge,
		"*",
		"*",
		true,
		true,
	)

	return nil
}
