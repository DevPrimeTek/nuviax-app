// Package crypto tests — unit tests for password hashing, SHA-256 digests,
// AES-256-GCM envelope encryption and random helpers.
package crypto

import (
	"strings"
	"testing"
)

// ── Bcrypt ────────────────────────────────────────────────────────────────────

func TestHashPassword_CheckRoundtrip(t *testing.T) {
	hash, err := HashPassword("CorrectHorseBatteryStaple!1")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if !CheckPassword("CorrectHorseBatteryStaple!1", hash) {
		t.Fatal("CheckPassword should accept the original password")
	}
	if CheckPassword("wrong-password", hash) {
		t.Fatal("CheckPassword must reject a different password")
	}
}

func TestHashPassword_ProducesDifferentHashes(t *testing.T) {
	h1, _ := HashPassword("samePass123!")
	h2, _ := HashPassword("samePass123!")
	if h1 == h2 {
		t.Fatal("bcrypt hashes should include unique salts — hashes must differ")
	}
	if !CheckPassword("samePass123!", h1) || !CheckPassword("samePass123!", h2) {
		t.Fatal("both hashes must validate against the original password")
	}
}

// ── SHA-256 ───────────────────────────────────────────────────────────────────

func TestSHA256Hex_Deterministic(t *testing.T) {
	a := SHA256Hex("nuviax")
	b := SHA256Hex("nuviax")
	if a != b {
		t.Fatalf("SHA256Hex must be deterministic, got %s vs %s", a, b)
	}
	if len(a) != 64 {
		t.Fatalf("expected 64 hex chars, got %d (%s)", len(a), a)
	}
}

func TestSHA256Hex_DifferentInputs(t *testing.T) {
	if SHA256Hex("a") == SHA256Hex("b") {
		t.Fatal("different inputs must produce different SHA-256 digests")
	}
}

// ── AES-256-GCM ───────────────────────────────────────────────────────────────

func testKey32() []byte {
	return []byte("0123456789abcdef0123456789abcdef") // 32 bytes
}

func TestAES_EncryptDecryptRoundtrip(t *testing.T) {
	key := testKey32()
	plain := "token=abc123;user=42"

	ct, err := Encrypt(plain, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if ct == "" || strings.Contains(ct, plain) {
		t.Fatalf("ciphertext looks wrong: %q", ct)
	}

	got, err := Decrypt(ct, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if got != plain {
		t.Fatalf("expected %q, got %q", plain, got)
	}
}

func TestAES_RejectsWrongKeySize(t *testing.T) {
	short := []byte("too-short")
	if _, err := Encrypt("x", short); err == nil {
		t.Fatal("Encrypt must reject keys != 32 bytes")
	}
	if _, err := Decrypt("AAAA", short); err == nil {
		t.Fatal("Decrypt must reject keys != 32 bytes")
	}
}

func TestAES_DecryptWithWrongKeyFails(t *testing.T) {
	k1 := testKey32()
	k2 := []byte("abcdefghijklmnopqrstuvwxyz012345") // different 32-byte key
	ct, err := Encrypt("hello", k1)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if _, err := Decrypt(ct, k2); err == nil {
		t.Fatal("Decrypt must fail with a different key (GCM auth tag mismatch)")
	}
}

func TestAES_DecryptRejectsShortCiphertext(t *testing.T) {
	key := testKey32()
	if _, err := Decrypt("AAAA", key); err == nil {
		t.Fatal("Decrypt must reject ciphertext shorter than nonce size")
	}
}

// ── Random helpers ────────────────────────────────────────────────────────────

func TestRandomBytes_LengthAndUniqueness(t *testing.T) {
	a, err := RandomBytes(32)
	if err != nil {
		t.Fatalf("RandomBytes failed: %v", err)
	}
	if len(a) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(a))
	}
	b, _ := RandomBytes(32)
	if string(a) == string(b) {
		t.Fatal("two consecutive RandomBytes calls should differ")
	}
}

func TestRandomHex_EncodesToHex(t *testing.T) {
	s, err := RandomHex(16)
	if err != nil {
		t.Fatalf("RandomHex failed: %v", err)
	}
	if len(s) != 32 {
		t.Fatalf("expected 32 hex chars from 16 bytes, got %d (%s)", len(s), s)
	}
	for _, c := range s {
		if !strings.ContainsRune("0123456789abcdef", c) {
			t.Fatalf("non-hex character in output: %q", s)
		}
	}
}
