package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("your_secret_key")

type SimFoxClaims struct {
	TokenKind string `json:"kind,omitempty"`
	WorkerID  uint   `json:"workerId,omitempty"`
	TenantID  uint   `json:"tenantId,omitempty"`
	jwt.RegisteredClaims
}

func GenerateToken(username string) (string, error) {
	claims := &SimFoxClaims{
		TokenKind: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func GenerateWorkerToken(workerName string, workerID uint, tenantID uint) (string, error) {
	claims := &SimFoxClaims{
		TokenKind: "worker",
		WorkerID:  workerID,
		TenantID:  tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   workerName,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ParseToken(tokenStr string) (string, error) {
	claims := &SimFoxClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return "", err
	}
	if claims.TokenKind != "" && claims.TokenKind != "user" {
		return "", jwt.ErrTokenInvalidClaims
	}
	return claims.Subject, nil
}

func ParseWorkerToken(tokenStr string) (string, uint, uint, error) {
	claims := &SimFoxClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return "", 0, 0, err
	}
	if claims.TokenKind != "worker" {
		return "", 0, 0, jwt.ErrTokenInvalidClaims
	}
	return claims.Subject, claims.WorkerID, claims.TenantID, nil
}

func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
