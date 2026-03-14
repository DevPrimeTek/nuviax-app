package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/devprimetek/nuviax-app/pkg/crypto"
)

// ── JWT Claims ────────────────────────────────────────────────────────────────

type Claims struct {
	UserID string `json:"uid"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type Service struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewService(privateKeyB64, publicKeyB64 []byte) (*Service, error) {
	privPEM, err := base64.StdEncoding.DecodeString(string(privateKeyB64))
	if err != nil {
		// Try raw PEM
		privPEM = privateKeyB64
	}
	pubPEM, err2 := base64.StdEncoding.DecodeString(string(publicKeyB64))
	if err2 != nil {
		pubPEM = publicKeyB64
	}

	privBlock, _ := pem.Decode(privPEM)
	if privBlock == nil {
		return nil, errors.New("invalid private key PEM")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
	if err != nil {
		// Try PKCS8
		key, err2 := x509.ParsePKCS8PrivateKey(privBlock.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not an RSA private key")
		}
	}

	pubBlock, _ := pem.Decode(pubPEM)
	if pubBlock == nil {
		return nil, errors.New("invalid public key PEM")
	}
	pubIface, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	publicKey, ok := pubIface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return &Service{privateKey: privateKey, publicKey: publicKey}, nil
}

// ── Generate tokens ───────────────────────────────────────────────────────────

func (s *Service) GenerateAccessToken(userID uuid.UUID, email string) (string, error) {
	jti, _ := crypto.RandomHex(16)
	claims := Claims{
		UserID: userID.String(),
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			Issuer:    "nuviax-api",
			Subject:   userID.String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

func (s *Service) GenerateRefreshToken() (string, error) {
	return crypto.RandomHex(32)
}

// ── Parse & validate ──────────────────────────────────────────────────────────

func (s *Service) ParseAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.publicKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (s *Service) GetJTI(tokenString string) (string, error) {
	claims, err := s.ParseAccessToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.ID, nil
}

// ── TOTP (MFA) ────────────────────────────────────────────────────────────────

func GenerateTOTPSecret(accountName string) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      "NUViaX",
		AccountName: accountName,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
}

func ValidateTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}

// ── Grade mapping (opaque score → human label) ────────────────────────────────

func ScoreToGrade(score float64) (string, string) {
	switch {
	case score >= 0.90:
		return "A+", "Excepțional"
	case score >= 0.80:
		return "A", "Excelent"
	case score >= 0.70:
		return "B", "Bun"
	case score >= 0.60:
		return "C", "Acceptabil"
	default:
		return "D", "Necesită atenție"
	}
}

// Localized grade label
func GradeLabel(grade, locale string) string {
	labels := map[string]map[string]string{
		"A+": {"ro": "Excepțional", "en": "Exceptional", "ru": "Исключительно"},
		"A":  {"ro": "Excelent",    "en": "Excellent",   "ru": "Отлично"},
		"B":  {"ro": "Bun",         "en": "Good",        "ru": "Хорошо"},
		"C":  {"ro": "Acceptabil",  "en": "Acceptable",  "ru": "Приемлемо"},
		"D":  {"ro": "Necesită atenție", "en": "Needs attention", "ru": "Требует внимания"},
	}
	if l, ok := labels[grade]; ok {
		if v, ok := l[locale]; ok {
			return v
		}
		return l["en"]
	}
	return grade
}
