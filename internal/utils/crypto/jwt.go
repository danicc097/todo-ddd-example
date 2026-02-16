package crypto

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthClaims struct {
	jwt.RegisteredClaims

	UserID      uuid.UUID `json:"uid"`
	MFAVerified bool      `json:"mfa"`
}

// TokenIssuer handles signing JWTs using a private RSA key.
type TokenIssuer struct {
	privateKey *rsa.PrivateKey
	issuer     string
}

func NewTokenIssuer(privateKey *rsa.PrivateKey, issuer string) *TokenIssuer {
	return &TokenIssuer{privateKey: privateKey, issuer: issuer}
}

func (i *TokenIssuer) Issue(userID uuid.UUID, mfaVerified bool, duration time.Duration) (string, error) {
	claims := AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    i.issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:      userID,
		MFAVerified: mfaVerified,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	return token.SignedString(i.privateKey)
}

// TokenVerifier handles verifying JWTs using a public RSA key.
type TokenVerifier struct {
	publicKey *rsa.PublicKey
}

func NewTokenVerifier(publicKey *rsa.PublicKey) *TokenVerifier {
	return &TokenVerifier{publicKey: publicKey}
}

func (v *TokenVerifier) Verify(tokenString string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return v.publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AuthClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
