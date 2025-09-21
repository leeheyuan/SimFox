package utils

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateRandomString(n int) (string, error) {
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// 使用 base64 编码并移除非字母数字字符
	str := base64.URLEncoding.EncodeToString(bytes)
	return str[:n], nil
}
