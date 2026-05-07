package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims

	UserID string `json:"user_id"`
}

type HS256Signer struct {
	Secret     []byte
	Issuer     string
	Audience   string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func generateJTI() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)

	return hex.EncodeToString(b)
}

func (s HS256Signer) NewAccessToken(userID string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.Issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{s.Audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(s.AccessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        generateJTI(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["typ"] = "JWT"

	return token.SignedString(s.Secret)
}

func (s HS256Signer) NewRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    s.Issuer,
		Subject:   userID,
		Audience:  jwt.ClaimStrings{s.Audience},
		ExpiresAt: jwt.NewNumericDate(now.Add(s.RefreshTTL)),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        generateJTI(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["typ"] = "JWT"

	return token.SignedString(s.Secret)
}

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
