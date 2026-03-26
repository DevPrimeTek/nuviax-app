// Package email — Resend.com transactional email integration for NuviaX
//
// Uses direct HTTP calls to the Resend API.
// No SDK required — only stdlib net/http.
// Requires RESEND_API_KEY environment variable.
// EMAIL_FROM defaults to "noreply@nuviax.app".
//
// Cost: ~$0/month on free tier (3,000 emails/month).
// Graceful degradation: if RESEND_API_KEY is absent, methods log and return nil.
package email

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	resendURL   = "https://api.resend.com/emails"
	httpTimeout = 10 * time.Second
)

// Client wraps the Resend email API.
type Client struct {
	apiKey string
	from   string
	http   *http.Client
}

// New creates an email client using RESEND_API_KEY from the environment.
// Returns nil and an error if the key is not set.
func New() (*Client, error) {
	key := os.Getenv("RESEND_API_KEY")
	if key == "" {
		return nil, errors.New("RESEND_API_KEY nu este configurat")
	}
	from := os.Getenv("EMAIL_FROM")
	if from == "" {
		from = "NuviaX <noreply@nuviax.app>"
	}
	return &Client{
		apiKey: key,
		from:   from,
		http:   &http.Client{Timeout: httpTimeout},
	}, nil
}

// IsAvailable returns true if RESEND_API_KEY is set.
func IsAvailable() bool {
	return os.Getenv("RESEND_API_KEY") != ""
}

// ── Resend API structs ────────────────────────────────────────────────────────

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
}

type sendResponse struct {
	ID    string `json:"id"`
	Error *struct {
		Name    string `json:"name"`
		Message string `json:"message"`
	} `json:"error"`
}

// send delivers a single email via Resend API.
func (c *Client) send(ctx context.Context, to, subject, html string) error {
	body := sendRequest{
		From:    c.from,
		To:      []string{to},
		Subject: subject,
		Html:    html,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", resendURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend api error %d: %s", resp.StatusCode, string(raw))
	}

	var result sendResponse
	if err := json.Unmarshal(raw, &result); err == nil && result.Error != nil {
		return fmt.Errorf("resend error %s: %s", result.Error.Name, result.Error.Message)
	}
	return nil
}

// ── Public methods ────────────────────────────────────────────────────────────

// SendWelcome sends a welcome email after successful registration.
func (c *Client) SendWelcome(ctx context.Context, to, name string) error {
	if name == "" {
		name = "utilizator"
	}
	return c.send(ctx, to, "Bun venit la NuviaX! 🎯", welcomeHTML(name))
}

// SendPasswordReset sends a password reset link valid for 1 hour.
func (c *Client) SendPasswordReset(ctx context.Context, to, resetLink string) error {
	return c.send(ctx, to, "Resetare parolă — NuviaX", resetHTML(resetLink))
}

// SendSprintComplete notifies the user when a 30-day sprint is closed.
func (c *Client) SendSprintComplete(ctx context.Context, to, name, goalName, grade string, sprintNum int) error {
	subject := fmt.Sprintf("Etapa %d completată — NuviaX", sprintNum)
	return c.send(ctx, to, subject, sprintCompleteHTML(name, goalName, grade, sprintNum))
}

// ── HTML Templates ────────────────────────────────────────────────────────────

func welcomeHTML(name string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="ro">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head>
<body style="margin:0;padding:0;background:#0a0a0a;font-family:'DM Sans',Arial,sans-serif;color:#e5e5e5">
<div style="max-width:560px;margin:40px auto;background:#111;border:1px solid #222;border-radius:16px;overflow:hidden">
  <div style="background:linear-gradient(135deg,#f97316,#ea580c);padding:32px;text-align:center">
    <h1 style="margin:0;color:#fff;font-size:28px;font-weight:700">NuviaX</h1>
    <p style="margin:8px 0 0;color:rgba(255,255,255,.85);font-size:14px">Growth Framework REV 5.6</p>
  </div>
  <div style="padding:32px">
    <h2 style="margin:0 0 16px;color:#f5f5f5;font-size:20px">Bun venit, %s!</h2>
    <p style="margin:0 0 16px;color:#a3a3a3;line-height:1.6">
      Contul tău NuviaX a fost creat cu succes. Platforma îți oferă un sistem complet
      de management al obiectivelor bazat pe framework-ul NUViaX REV 5.6 cu 40 de componente.
    </p>
    <p style="margin:0 0 24px;color:#a3a3a3;line-height:1.6">
      Începe prin a-ți defini primul obiectiv principal (GO) și lasă framework-ul să îți
      genereze planul de sprint de 30 de zile.
    </p>
    <a href="https://nuviax.app/onboarding"
       style="display:inline-block;background:#f97316;color:#fff;padding:14px 28px;border-radius:10px;text-decoration:none;font-weight:600;font-size:15px">
      Pornește onboarding-ul →
    </a>
  </div>
  <div style="padding:20px 32px;border-top:1px solid #222;text-align:center">
    <p style="margin:0;color:#525252;font-size:12px">
      © 2026 DevPrimeTek · <a href="https://nuviax.app" style="color:#f97316;text-decoration:none">nuviax.app</a>
    </p>
  </div>
</div>
</body></html>`, name)
}

func resetHTML(resetLink string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="ro">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head>
<body style="margin:0;padding:0;background:#0a0a0a;font-family:'DM Sans',Arial,sans-serif;color:#e5e5e5">
<div style="max-width:560px;margin:40px auto;background:#111;border:1px solid #222;border-radius:16px;overflow:hidden">
  <div style="background:linear-gradient(135deg,#f97316,#ea580c);padding:32px;text-align:center">
    <h1 style="margin:0;color:#fff;font-size:28px;font-weight:700">NuviaX</h1>
  </div>
  <div style="padding:32px">
    <h2 style="margin:0 0 16px;color:#f5f5f5;font-size:20px">Resetare parolă</h2>
    <p style="margin:0 0 16px;color:#a3a3a3;line-height:1.6">
      Ai solicitat resetarea parolei pentru contul tău NuviaX.
      Apasă butonul de mai jos pentru a seta o parolă nouă.
    </p>
    <p style="margin:0 0 24px;color:#737373;font-size:13px">
      Link-ul este valabil <strong style="color:#f5f5f5">1 oră</strong>.
      Dacă nu ai solicitat tu resetarea, ignoră acest email.
    </p>
    <a href="%s"
       style="display:inline-block;background:#f97316;color:#fff;padding:14px 28px;border-radius:10px;text-decoration:none;font-weight:600;font-size:15px">
      Resetează parola →
    </a>
    <p style="margin:24px 0 0;color:#525252;font-size:12px;word-break:break-all">
      Sau copiază acest link: %s
    </p>
  </div>
  <div style="padding:20px 32px;border-top:1px solid #222;text-align:center">
    <p style="margin:0;color:#525252;font-size:12px">
      © 2026 DevPrimeTek · <a href="https://nuviax.app" style="color:#f97316;text-decoration:none">nuviax.app</a>
    </p>
  </div>
</div>
</body></html>`, resetLink, resetLink)
}

func sprintCompleteHTML(name, goalName, grade string, sprintNum int) string {
	gradeColor := map[string]string{
		"A": "#22c55e", "B": "#3b82f6", "C": "#f97316", "D": "#ef4444",
	}
	color := gradeColor[grade]
	if color == "" {
		color = "#f97316"
	}
	gradeLabel := map[string]string{
		"A": "Excelent", "B": "Bun", "C": "Satisfăcător", "D": "Slab",
	}
	label := gradeLabel[grade]
	if label == "" {
		label = grade
	}
	if name == "" {
		name = "utilizator"
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="ro">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head>
<body style="margin:0;padding:0;background:#0a0a0a;font-family:'DM Sans',Arial,sans-serif;color:#e5e5e5">
<div style="max-width:560px;margin:40px auto;background:#111;border:1px solid #222;border-radius:16px;overflow:hidden">
  <div style="background:linear-gradient(135deg,#f97316,#ea580c);padding:32px;text-align:center">
    <h1 style="margin:0;color:#fff;font-size:28px;font-weight:700">NuviaX</h1>
    <p style="margin:8px 0 0;color:rgba(255,255,255,.85);font-size:14px">Etapa %d completată</p>
  </div>
  <div style="padding:32px">
    <h2 style="margin:0 0 8px;color:#f5f5f5;font-size:20px">Felicitări, %s!</h2>
    <p style="margin:0 0 24px;color:#a3a3a3;line-height:1.6">
      Ai finalizat <strong style="color:#f5f5f5">Etapa %d</strong> pentru obiectivul
      <strong style="color:#f97316">%s</strong>.
    </p>
    <div style="background:#1a1a1a;border:1px solid #2a2a2a;border-radius:12px;padding:20px;text-align:center;margin-bottom:24px">
      <div style="font-size:48px;font-weight:700;color:%s;margin-bottom:4px">%s</div>
      <div style="color:#737373;font-size:14px">%s</div>
    </div>
    <p style="margin:0 0 24px;color:#a3a3a3;line-height:1.6">
      Verifică recap-ul complet al etapei și pregătește-te pentru
      <strong style="color:#f5f5f5">Etapa %d</strong>.
    </p>
    <a href="https://nuviax.app/recap"
       style="display:inline-block;background:#f97316;color:#fff;padding:14px 28px;border-radius:10px;text-decoration:none;font-weight:600;font-size:15px">
      Vezi recap-ul →
    </a>
  </div>
  <div style="padding:20px 32px;border-top:1px solid #222;text-align:center">
    <p style="margin:0;color:#525252;font-size:12px">
      © 2026 DevPrimeTek · <a href="https://nuviax.app" style="color:#f97316;text-decoration:none">nuviax.app</a>
    </p>
  </div>
</div>
</body></html>`, sprintNum, name, sprintNum, goalName, color, grade, label, sprintNum+1)
}
