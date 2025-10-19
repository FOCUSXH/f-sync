package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
)

// HashFile 计算指定文件的哈希值，使用给定的哈希函数构造器。
// 如果 hasher 为 nil，则默认使用 sha256.New()。
func HashFile(filePath string, hasher func() hash.Hash) (string, error) {
	if hasher == nil {
		hasher = sha256.New
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("无法打开文件 %s: %w", filePath, err)
	}
	defer file.Close()

	h := hasher()
	if _, err := io.Copy(h, file); err != nil {
		return "", fmt.Errorf("读取文件内容计算哈希失败: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// HashFileSHA256 是 HashFile 的便捷封装，固定使用 SHA256。
func HashFileSHA256(filePath string) (string, error) {
	return HashFile(filePath, sha256.New)
}
