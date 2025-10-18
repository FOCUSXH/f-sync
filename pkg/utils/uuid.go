// pkg/utils/uuid.go
package utils

import "github.com/google/uuid"

// NewUUID 生成一个随机的 UUID v4 字符串
func NewUUID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

// MustNewUUID 生成 UUID，失败时 panic（适用于初始化等关键场景）
func MustNewUUID() string {
	id, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return id.String()
}