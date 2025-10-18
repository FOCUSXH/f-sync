package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/argon2"
)

const (
	saltLen   = 16
	nonceLen  = 12
	keyLen    = 32 // AES-256
	timeCost  = 1
	memoryKiB = 64 * 1024
	threads   = 4
)

// EncryptFile encrypts a file using the given password and salt.
// The output file format is: [salt][nonce][ciphertext+tag]
func EncryptFile(inputPath, outputPath string, password string, salt []byte) error {
	if len(salt) != saltLen {
		return fmt.Errorf("invalid salt length: expected %d bytes, got %d", saltLen, len(salt))
	}

	// Read plaintext
	plaintext, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Derive key from password and salt
	key := argon2.IDKey([]byte(password), salt, timeCost, memoryKiB, threads, keyLen)

	// AES-GCM setup
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil) // includes auth tag

	// Write output: salt + nonce + ciphertext
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if _, err := outFile.Write(salt); err != nil {
		return fmt.Errorf("failed to write salt: %w", err)
	}
	if _, err := outFile.Write(nonce); err != nil {
		return fmt.Errorf("failed to write nonce: %w", err)
	}
	if _, err := outFile.Write(ciphertext); err != nil {
		return fmt.Errorf("failed to write ciphertext: %w", err)
	}

	return nil
}

// DecryptFile decrypts a file and verifies that the salt in the file matches expectedSalt.
// The input file format must be: [salt][nonce][ciphertext+tag]
func DecryptFile(inputPath, outputPath string, password string, expectedSalt []byte) error {
	if len(expectedSalt) != saltLen {
		return fmt.Errorf("invalid expectedSalt length: expected %d bytes, got %d", saltLen, len(expectedSalt))
	}

	// Read entire encrypted file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read encrypted file: %w", err)
	}

	if len(data) < saltLen+nonceLen+1 {
		return fmt.Errorf("encrypted file is too short to be valid")
	}

	// Extract salt from file
	fileSalt := data[:saltLen]

	// ✅ Verify salt matches expectedSalt
	if !equal(fileSalt, expectedSalt) {
		return fmt.Errorf("salt mismatch: file contains different salt (possible wrong password or corrupted file)")
	}

	// Extract nonce and ciphertext
	nonce := data[saltLen : saltLen+nonceLen]
	ciphertext := data[saltLen+nonceLen:]

	// Derive key using expectedSalt (same as fileSalt)
	key := argon2.IDKey([]byte(password), expectedSalt, timeCost, memoryKiB, threads, keyLen)

	// AES-GCM setup
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("decryption failed (wrong password or tampered file): %w", err)
	}

	// Write plaintext
	if err := os.WriteFile(outputPath, plaintext, 0600); err != nil {
		return fmt.Errorf("failed to write decrypted file: %w", err)
	}

	return nil
}

// equal compares two byte slices in constant time to avoid timing attacks.
func equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	result := byte(0)
	for i := range a {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
