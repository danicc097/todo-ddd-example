package crypto

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	tokenAudience = "todo-ddd-api-clients"
)

type AuthClaims struct {
	jwt.RegisteredClaims

	UserID      uuid.UUID `json:"uid"`
	MFAVerified bool      `json:"mfa"`
}

type TokenProvider struct {
	Issuer   *TokenIssuer
	Verifier *TokenVerifier
}

func NewTokenProvider(privKeyPath, pubKeyPath, issuerName string) (*TokenProvider, error) {
	privKeyBytes, err := os.ReadFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	pubKeyBytes, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return &TokenProvider{
		Issuer:   NewTokenIssuer(privKey, issuerName),
		Verifier: NewTokenVerifier(pubKey),
	}, nil
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
	now := time.Now()
	claims := AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    i.issuer,
			Audience:  jwt.ClaimStrings{tokenAudience},
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
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
