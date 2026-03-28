package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ParseOptions struct {
	ExpectedIssuer   string
	ExpectedAudience string
	AllowedMethods   []string
	Leeway           time.Duration
}

func (s HS256Signer) ParseAccessToken(tokenStr string, opt ParseOptions) (*Claims, error) {
	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected alg: %v", t.Header["alg"])
			}

			return s.Secret, nil
		},
		jwt.WithIssuer(opt.ExpectedIssuer),
		jwt.WithAudience(opt.ExpectedAudience),
		jwt.WithValidMethods(opt.AllowedMethods),
		jwt.WithLeeway(opt.Leeway),
	)

	if err != nil {
		return nil, err
	}

	if !tkn.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (s HS256Signer) ParseRefreshToken(tokenStr string, opt ParseOptions) (*jwt.RegisteredClaims, error) {
	claims := &jwt.RegisteredClaims{}

	tkn, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(_ *jwt.Token) (any, error) {
			return s.Secret, nil
		},
		jwt.WithIssuer(opt.ExpectedIssuer),
		jwt.WithAudience(opt.ExpectedAudience),
		jwt.WithValidMethods(opt.AllowedMethods),
		jwt.WithLeeway(opt.Leeway),
	)

	if err != nil {
		return nil, err
	}

	if !tkn.Valid {
		return nil, errors.New("refresh is invalid")
	}

	return claims, nil
}
