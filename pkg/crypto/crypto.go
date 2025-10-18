package crypto

import (
	"crypto/rand"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

func HashStringByBcrypt(plaintext string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func VerifyStringWithBcrypt(hashed, plaintext string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plaintext))
	return err == nil
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	return salt, err
}



// ReDeriveKey 使用已知 salt 和密码重新派生密钥（用于解密）
func ReDeriveKey(password string, salt []byte, keyLen int) []byte {
	const (
		time    = 1
		memory  = 64 * 1024
		threads = 4
	)
	return argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(keyLen))
}
