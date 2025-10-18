package utils

import (
	"fmt"
	"unicode"
)

// ValidatePassword 验证密码强度
// 密码必须满足以下条件：
// 1. 长度至少8位
// 2. 包含字母
// 3. 包含数字
// 4. 包含特殊符号
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("密码长度必须至少8位")
	}

	var (
		hasLetter  bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsLetter(char):
			hasLetter = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasLetter {
		return fmt.Errorf("密码必须包含字母")
	}

	if !hasDigit {
		return fmt.Errorf("密码必须包含数字")
	}

	if !hasSpecial {
		return fmt.Errorf("密码必须包含特殊符号")
	}

	return nil
}

