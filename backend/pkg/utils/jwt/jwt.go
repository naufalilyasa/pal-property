package jwt

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
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
		"exp": time.Now().Add(time.Duration(config.Env.JwtAccessExpiration) * time.Second).Unix(),
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
		"exp": time.Now().Add(time.Duration(config.Env.JwtRefreshExpiration) * time.Second).Unix(),
		"iat": time.Now().Unix(),
	}
	refToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshTokenClaims)
	refreshToken, err = refToken.SignedString(privateKey)
	if err != nil {
		return "", "", "", err
	}

	return accessToken, refreshToken, jti, nil
}

// ValidateAccessToken parses the access token and returns the user ID.
func ValidateAccessToken(tokenString string) (uuid.UUID, error) {
	if publicKey == nil {
		if err := InitKeys(); err != nil {
			return uuid.Nil, err
		}
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signature method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return publicKey, nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		subStr, ok := claims["sub"].(string)
		if !ok {
			return uuid.Nil, errors.New("invalid 'sub' claim in token")
		}
		userID, err := uuid.Parse(subStr)
		if err != nil {
			return uuid.Nil, errors.New("invalid user ID format in token")
		}
		return userID, nil
	}

	return uuid.Nil, errors.New("invalid token claims")
}

// ValidateRefreshToken parses the refresh token and returns the user ID and jti.
func ValidateRefreshToken(tokenString string) (uuid.UUID, string, error) {
	if publicKey == nil {
		if err := InitKeys(); err != nil {
			return uuid.Nil, "", err
		}
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signature method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return publicKey, nil
	})

	if err != nil {
		return uuid.Nil, "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		subStr, ok := claims["sub"].(string)
		if !ok {
			return uuid.Nil, "", errors.New("invalid 'sub' claim in refresh token")
		}
		userID, err := uuid.Parse(subStr)
		if err != nil {
			return uuid.Nil, "", errors.New("invalid user ID format in refresh token")
		}

		jti, ok := claims["jti"].(string)
		if !ok {
			return uuid.Nil, "", errors.New("invalid 'jti' claim in refresh token")
		}

		return userID, jti, nil
	}

	return uuid.Nil, "", errors.New("invalid refresh token claims")
}
