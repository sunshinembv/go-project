package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testSigner() HS256Signer {
	return HS256Signer{
		Secret:     []byte("Secret123321"),
		Issuer:     "todo_list-service",
		Audience:   "todo_list-client",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}
}

func testParseOptions(signer HS256Signer) ParseOptions {
	return ParseOptions{
		ExpectedIssuer:   signer.Issuer,
		ExpectedAudience: signer.Audience,
		AllowedMethods:   []string{"HS256"},
		Leeway:           time.Second,
	}
}

func TestGenerateJTI(t *testing.T) {
	first := generateJTI()
	second := generateJTI()

	assert.Len(t, first, 32)
	assert.Len(t, second, 32)
	assert.NotEqual(t, first, second)
}

func TestAccessToken(t *testing.T) {
	signer := testSigner()

	token, err := signer.NewAccessToken("user-123")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := signer.ParseAccessToken(token, testParseOptions(signer))
	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "user-123", claims.Subject)
	assert.Equal(t, signer.Issuer, claims.Issuer)
	assert.Contains(t, claims.Audience, signer.Audience)
	assert.NotEmpty(t, claims.ID)
	assert.WithinDuration(t, time.Now(), claims.IssuedAt.Time, 2*time.Second)
	assert.WithinDuration(t, time.Now().Add(signer.AccessTTL), claims.ExpiresAt.Time, 2*time.Second)
}

func TestRefreshToken(t *testing.T) {
	signer := testSigner()

	token, err := signer.NewRefreshToken("user-123")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := signer.ParseRefreshToken(token, testParseOptions(signer))
	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, "user-123", claims.Subject)
	assert.Equal(t, signer.Issuer, claims.Issuer)
	assert.Contains(t, claims.Audience, signer.Audience)
	assert.NotEmpty(t, claims.ID)
}

func TestParseAccessTokenErrors(t *testing.T) {
	signer := testSigner()
	validToken, err := signer.NewAccessToken("user-123")
	require.NoError(t, err)

	expiredSigner := signer
	expiredSigner.AccessTTL = -time.Hour
	expiredToken, err := expiredSigner.NewAccessToken("user-123")
	require.NoError(t, err)

	noneToken, err := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{
		UserID: "user-123",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    signer.Issuer,
			Audience:  jwt.ClaimStrings{signer.Audience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	type test struct {
		name   string
		token  string
		signer HS256Signer
		opt    ParseOptions
	}

	tests := []test{
		{
			name:   "malformed",
			token:  "invalid",
			signer: signer,
			opt:    testParseOptions(signer),
		},
		{
			name:  "wrong secret",
			token: validToken,
			signer: HS256Signer{
				Secret: []byte("wrong"),
			},
			opt: testParseOptions(signer),
		},
		{
			name:   "wrong issuer",
			token:  validToken,
			signer: signer,
			opt: ParseOptions{
				ExpectedIssuer:   "wrong",
				ExpectedAudience: signer.Audience,
				AllowedMethods:   []string{"HS256"},
			},
		},
		{
			name:   "wrong audience",
			token:  validToken,
			signer: signer,
			opt: ParseOptions{
				ExpectedIssuer:   signer.Issuer,
				ExpectedAudience: "wrong",
				AllowedMethods:   []string{"HS256"},
			},
		},
		{
			name:   "expired",
			token:  expiredToken,
			signer: signer,
			opt:    testParseOptions(signer),
		},
		{
			name:   "unexpected algorithm",
			token:  noneToken,
			signer: signer,
			opt: ParseOptions{
				ExpectedIssuer:   signer.Issuer,
				ExpectedAudience: signer.Audience,
				AllowedMethods:   []string{"none"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := tc.signer.ParseAccessToken(tc.token, tc.opt)
			require.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

func TestParseRefreshTokenErrors(t *testing.T) {
	signer := testSigner()
	validToken, err := signer.NewRefreshToken("user-123")
	require.NoError(t, err)

	type test struct {
		name   string
		token  string
		signer HS256Signer
		opt    ParseOptions
	}

	tests := []test{
		{
			name:   "malformed",
			token:  "invalid",
			signer: signer,
			opt:    testParseOptions(signer),
		},
		{
			name:  "wrong secret",
			token: validToken,
			signer: HS256Signer{
				Secret: []byte("wrong"),
			},
			opt: testParseOptions(signer),
		},
		{
			name:   "wrong method",
			token:  validToken,
			signer: signer,
			opt: ParseOptions{
				ExpectedIssuer:   signer.Issuer,
				ExpectedAudience: signer.Audience,
				AllowedMethods:   []string{"HS512"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := tc.signer.ParseRefreshToken(tc.token, tc.opt)
			require.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}
