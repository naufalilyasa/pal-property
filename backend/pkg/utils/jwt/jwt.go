package jwt

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/username/pal-property-backend/pkg/config"
)

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)

// InitKeys initializes the RSA keys from base64 strings in the configuration.
func InitKeys() error {
	if config.Env.JwtPrivateKeyBase64 == "" || config.Env.JwtPublicKeyBase64 == "" {
		return errors.New("JWT base64 keys are not set in config")
	}

	// Parse Private Key
	privKeyBytes, err := base64.StdEncoding.DecodeString(config.Env.JwtPrivateKeyBase64)
	if err != nil {
		return err
	}
	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privKeyBytes)
	if err != nil {
		return err
	}

	// Parse Public Key
	pubKeyBytes, err := base64.StdEncoding.DecodeString(config.Env.JwtPublicKeyBase64)
	if err != nil {
		return err
	}
	publicKey, err = jwt.ParseRSAPublicKeyFromPEM(pubKeyBytes)
	if err != nil {
		return err
	}

	return nil
}

// GenerateTokens creates both an access token and a refresh token for a user.
func GenerateTokens(userID uuid.UUID) (accessToken string, refreshToken string, jti string, err error) {
	if privateKey == nil {
		if err := InitKeys(); err != nil {
			return "", "", "", err
		}
	}

	// 1. Generate Access Token
	accessTokenClaims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(config.Env.JwtAccessExpiration).Unix(),
		"iat": time.Now().Unix(),
	}
	accToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessTokenClaims)
	accessToken, err = accToken.SignedString(privateKey)
	if err != nil {
		return "", "", "", err
	}

	// 2. Generate Refresh Token
	jtiUUID, err := uuid.NewV7()
	if err != nil {
		return "", "", "", err
	}
	jti = jtiUUID.String()

	refreshTokenClaims := jwt.MapClaims{
		"sub": userID.String(),
		"jti": jti,
		"exp": time.Now().Add(config.Env.JwtRefreshExpiration).Unix(),
		"iat": time.Now().Unix(),
	}
	refToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshTokenClaims)
	refreshToken, err = refToken.SignedString(privateKey)
	if err != nil {
		return "", "", "", err
	}

	return accessToken, refreshToken, jti, nil
}
