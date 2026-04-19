// Package email tests — unit tests for the Resend transactional layer.
//
// These tests only exercise pure/stateless helpers (no live Resend calls).
// For live-delivery validation see docs/testing/smoke-test-report.md §Email.
package email

import (
	"os"
	"strings"
	"testing"
)

// ── IsAvailable graceful-degradation check ────────────────────────────────────

func TestIsAvailable_MissingKey(t *testing.T) {
	prev := os.Getenv("RESEND_API_KEY")
	t.Cleanup(func() { _ = os.Setenv("RESEND_API_KEY", prev) })

	_ = os.Unsetenv("RESEND_API_KEY")
	if IsAvailable() {
		t.Fatal("IsAvailable() must return false when RESEND_API_KEY is unset")
	}
}

func TestIsAvailable_WithKey(t *testing.T) {
	prev := os.Getenv("RESEND_API_KEY")
	t.Cleanup(func() { _ = os.Setenv("RESEND_API_KEY", prev) })

	_ = os.Setenv("RESEND_API_KEY", "re_dummy_test_key")
	if !IsAvailable() {
		t.Fatal("IsAvailable() must return true when RESEND_API_KEY is set")
	}
}

// ── Client construction ──────────────────────────────────────────────────────

func TestNew_RequiresKey(t *testing.T) {
	prev := os.Getenv("RESEND_API_KEY")
	t.Cleanup(func() { _ = os.Setenv("RESEND_API_KEY", prev) })

	_ = os.Unsetenv("RESEND_API_KEY")
	if _, err := New(); err == nil {
		t.Fatal("expected error when RESEND_API_KEY is unset")
	}
}

func TestNew_UsesDefaultFrom(t *testing.T) {
	prevKey := os.Getenv("RESEND_API_KEY")
	prevFrom := os.Getenv("EMAIL_FROM")
	t.Cleanup(func() {
		_ = os.Setenv("RESEND_API_KEY", prevKey)
		_ = os.Setenv("EMAIL_FROM", prevFrom)
	})

	_ = os.Setenv("RESEND_API_KEY", "re_dummy_test_key")
	_ = os.Unsetenv("EMAIL_FROM")

	c, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.from != "NuviaX <noreply@nuviax.app>" {
		t.Fatalf("expected default FROM, got %q", c.from)
	}
}

func TestNew_UsesCustomFrom(t *testing.T) {
	prevKey := os.Getenv("RESEND_API_KEY")
	prevFrom := os.Getenv("EMAIL_FROM")
	t.Cleanup(func() {
		_ = os.Setenv("RESEND_API_KEY", prevKey)
		_ = os.Setenv("EMAIL_FROM", prevFrom)
	})

	_ = os.Setenv("RESEND_API_KEY", "re_dummy_test_key")
	_ = os.Setenv("EMAIL_FROM", "Support <support@example.com>")

	c, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.from != "Support <support@example.com>" {
		t.Fatalf("expected custom FROM, got %q", c.from)
	}
}

// ── HTML templates ────────────────────────────────────────────────────────────

func TestWelcomeHTML_ContainsName(t *testing.T) {
	html := welcomeHTML("Andrei")
	if !strings.Contains(html, "Andrei") {
		t.Fatal("welcome HTML must include the provided name")
	}
	if !strings.Contains(html, "NuviaX") {
		t.Fatal("welcome HTML must include the NuviaX brand")
	}
	if !strings.Contains(html, "https://nuviax.app/onboarding") {
		t.Fatal("welcome HTML must link to /onboarding")
	}
}

func TestResetHTML_ContainsLinkTwice(t *testing.T) {
	link := "https://nuviax.app/auth/reset-password?token=abc123"
	html := resetHTML(link)
	if strings.Count(html, link) < 2 {
		t.Fatal("reset HTML must embed the reset link at least twice (button + copy text)")
	}
	if !strings.Contains(html, "1 oră") {
		t.Fatal("reset HTML must state the 1-hour TTL")
	}
}

func TestSprintCompleteHTML_RendersGradeAndFallbacks(t *testing.T) {
	html := sprintCompleteHTML("Ana", "Lose 5kg", "A", 1)
	if !strings.Contains(html, "Ana") {
		t.Fatal("sprint-complete HTML must contain the user name")
	}
	if !strings.Contains(html, "Lose 5kg") {
		t.Fatal("sprint-complete HTML must contain the goal name")
	}
	if !strings.Contains(html, ">A<") {
		t.Fatal("sprint-complete HTML must render the grade letter")
	}
	if !strings.Contains(html, "Excelent") {
		t.Fatal("grade A should render the localised label 'Excelent'")
	}
	if !strings.Contains(html, "Etapa 2") {
		t.Fatal("sprint-complete HTML should announce the next sprint (N+1)")
	}
}

func TestSprintCompleteHTML_UnknownGradeFallsBack(t *testing.T) {
	html := sprintCompleteHTML("", "Goal", "Z", 5)
	if !strings.Contains(html, "utilizator") {
		t.Fatal("empty name should fall back to 'utilizator'")
	}
	if !strings.Contains(html, ">Z<") {
		t.Fatal("unknown grade should still render the raw letter")
	}
}
