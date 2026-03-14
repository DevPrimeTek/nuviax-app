package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 14

// ── Bcrypt ────────────────────────────────────────────────────────────────────

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// ── SHA-256 ───────────────────────────────────────────────────────────────────

func SHA256Hex(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])
}

// ── AES-256-GCM ───────────────────────────────────────────────────────────────

func Encrypt(plaintext string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", errors.New("encryption key must be 32 bytes (AES-256)")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(encoded string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", errors.New("encryption key must be 32 bytes (AES-256)")
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ns := gcm.NonceSize()
	if len(data) < ns {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:ns], data[ns:]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// ── Random ────────────────────────────────────────────────────────────────────

func RandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}

func RandomHex(n int) (string, error) {
	b, err := RandomBytes(n)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
