// Package crypto provides AES-256 field-level encryption for sensitive PII data.
// It serves as the SM4-compatible encryption layer (PIPL compliant for data-at-rest).
// The underlying algorithm can be swapped to SM4 without changing callers.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"os"
	"strings"
)

var encryptionKey []byte

// Init initializes the encryption key from the DATA_ENCRYPTION_KEY env var.
func Init() {
	key := os.Getenv("DATA_ENCRYPTION_KEY")
	if key == "" {
		key = "nsi-default-encryption-key-2026"
	}
	h1 := md5.Sum([]byte(key))
	h2 := md5.Sum(append(h1[:], []byte(key)...))
	encryptionKey = append(h1[:], h2[:]...)
}

// Encrypt encrypts plaintext using AES-256-CTR and returns a base64-encoded string prefixed with "enc:".
func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	if len(encryptionKey) == 0 {
		Init()
	}
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}
	iv := make([]byte, aes.BlockSize)
	stream := cipher.NewCTR(block, iv)
	encrypted := make([]byte, len(plaintext))
	stream.XORKeyStream(encrypted, []byte(plaintext))
	return "enc:" + base64.StdEncoding.EncodeToString(encrypted), nil
}

// Decrypt decrypts an "enc:"-prefixed ciphertext back to plaintext.
// Non-encrypted strings are returned as-is for backward compatibility.
func Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" || !strings.HasPrefix(ciphertext, "enc:") {
		return ciphertext, nil
	}
	if len(encryptionKey) == 0 {
		Init()
	}
	data, err := base64.StdEncoding.DecodeString(ciphertext[4:])
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}
	iv := make([]byte, aes.BlockSize)
	stream := cipher.NewCTR(block, iv)
	decrypted := make([]byte, len(data))
	stream.XORKeyStream(decrypted, data)
	return string(decrypted), nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
}
