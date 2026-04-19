// Package ai — Claude Haiku 4.5 integration for NuviaX
//
// Uses direct HTTP calls to the Anthropic Messages API.
// No SDK required — only stdlib net/http.
// Requires ANTHROPIC_API_KEY environment variable.
//
// Cost: ~$0.25/1M tokens input — estimated $4-5/month at 1K active users.
// Model: claude-haiku-4-5-20251001
package ai

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
	apiURL    = "https://api.anthropic.com/v1/messages"
	model     = "claude-haiku-4-5-20251001"
	apiVersion = "2023-06-01"
	// Timeout generously set so slow API doesn't block requests
	httpTimeout = 12 * time.Second
	// Max tokens for short responses (task names, analysis)
	maxTokensShort = 256
	maxTokensLong  = 512
)

// Client wraps the Anthropic Messages API.
type Client struct {
	apiKey string
	http   *http.Client
}

// New creates an AI client using ANTHROPIC_API_KEY from the environment.
// Returns nil and an error if the key is not set.
func New() (*Client, error) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		return nil, errors.New("ANTHROPIC_API_KEY nu este configurat")
	}
	return &Client{
		apiKey: key,
		http:   &http.Client{Timeout: httpTimeout},
	}, nil
}

// IsAvailable checks whether the ANTHROPIC_API_KEY is set.
// Use this for graceful degradation — if false, use rule-based fallbacks.
func IsAvailable() bool {
	return os.Getenv("ANTHROPIC_API_KEY") != ""
}

// ── Messages API structs ──────────────────────────────────────────────────────

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type requestBody struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []message `json:"messages"`
}

type responseBody struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// complete sends a single request to Claude and returns the text response.
func (c *Client) complete(ctx context.Context, system, userMsg string, maxTokens int) (string, error) {
	body := requestBody{
		Model:     model,
		MaxTokens: maxTokens,
		System:    system,
		Messages:  []message{{Role: "user", Content: userMsg}},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", apiVersion)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result responseBody
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", err
	}
	if result.Error != nil {
		return "", fmt.Errorf("anthropic api error: %s — %s", result.Error.Type, result.Error.Message)
	}
	if len(result.Content) == 0 {
		return "", errors.New("empty response from claude")
	}
	return result.Content[0].Text, nil
}

// ── Public methods ────────────────────────────────────────────────────────────

// GenerateTaskTexts generates contextualised daily task descriptions for a goal.
// Returns 1-3 task strings based on the goal name, checkpoint name, and sprint number.
// Falls back to empty slice on error (caller must use rule-based fallback).
func (c *Client) GenerateTaskTexts(ctx context.Context, goalName, checkpointName string, sprintNumber, count int) ([]string, error) {
	system := `Ești asistentul NuviaX — ajuți utilizatorii să-și atingă obiectivele de viață.
Generează sarcini zilnice concrete și acționabile, în română, scurte (max 10 cuvinte fiecare).
Răspunde DOAR cu lista de sarcini, una per linie, fără numerotare, fără explicații.`

	prompt := fmt.Sprintf(
		"Obiectiv: %q\nEtapă curentă: %d\nMilestone activ: %q\nGenerează %d sarcini zilnice concrete pentru azi:",
		goalName, sprintNumber, checkpointName, count,
	)

	text, err := c.complete(ctx, system, prompt, maxTokensShort)
	if err != nil {
		return nil, err
	}

	tasks := parseLines(text, count)
	return tasks, nil
}

// AnalyzeGO uses Claude to classify whether a goal is specific, measurable and time-bound.
// Returns (needsClarification bool, question string, hint string).
func (c *Client) AnalyzeGO(ctx context.Context, goalText string) (needsClarification bool, question, hint string, err error) {
	system := `Ești analistul NuviaX. Analizezi dacă un obiectiv de viață este SMART (Specific, Măsurabil, Realizabil, Relevant, Delimitat în timp).
Răspunde EXCLUSIV în format JSON valid cu câmpurile:
{
  "needs_clarification": true/false,
  "question": "întrebarea de clarificare (dacă needs_clarification=true, altfel null)",
  "hint": "exemplu concret (dacă needs_clarification=true, altfel null)"
}
Nu adăuga niciun text în afara JSON-ului.`

	prompt := fmt.Sprintf("Analizează acest obiectiv: %q", goalText)

	text, err := c.complete(ctx, system, prompt, maxTokensLong)
	if err != nil {
		return false, "", "", err
	}

	var result struct {
		NeedsClarification bool    `json:"needs_clarification"`
		Question           *string `json:"question"`
		Hint               *string `json:"hint"`
	}
	if jsonErr := json.Unmarshal([]byte(text), &result); jsonErr != nil {
		// If Claude returned invalid JSON, treat as needs clarification
		return true, "Poți descrie obiectivul mai specific? Ce rezultat concret vrei să obții și până când?", "", nil
	}

	q := ""
	if result.Question != nil {
		q = *result.Question
	}
	h := ""
	if result.Hint != nil {
		h = *result.Hint
	}
	return result.NeedsClarification, q, h, nil
}

// SuggestionResult holds the suggestion returned by SuggestGOCategory.
// Extended beyond a pure category to match the onboarding workflow
// (C2 Behavior Model + C9 Semantic Parsing): the AI also proposes a
// dominant behavior model and 1-3 alternative "directions" the user can
// pick from when the raw GO is ambiguous.
type SuggestionResult struct {
	Category      string   `json:"category"`
	Confidence    float64  `json:"confidence"`
	Reasoning     string   `json:"reasoning"`
	BehaviorModel string   `json:"behavior_model"`
	Directions    []string `json:"directions"`
}

// validBehaviorModels holds the canonical C2 set; used to sanity-check AI output.
var validBehaviorModels = map[string]bool{
	"CREATE": true, "INCREASE": true, "REDUCE": true,
	"MAINTAIN": true, "EVOLVE": true,
}

// SuggestGOCategory asks Claude Haiku to classify a goal AND propose:
//   - category (HEALTH/CAREER/FINANCE/RELATIONSHIPS/LEARNING/CREATIVITY/OTHER),
//   - dominant behavior model (C2: CREATE/INCREASE/REDUCE/MAINTAIN/EVOLVE),
//   - 1-3 alternative directions the user can pick from when the GO is
//     ambiguous (C9 Semantic Parsing).
//
// Hard timeout: 2 seconds — user is waiting in the onboarding flow.
// Returns an empty SuggestionResult (not an error) on timeout or API failure,
// so the caller can fall back to rule-based defaults.
func (c *Client) SuggestGOCategory(title, description string) SuggestionResult {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	system := `You are a NuviaX goal coach. Given a user goal (title + optional description), return a JSON analysis.

Respond ONLY with valid JSON — no prose, no markdown. Schema:
{
  "category": "HEALTH|CAREER|FINANCE|RELATIONSHIPS|LEARNING|CREATIVITY|OTHER",
  "confidence": 0.0-1.0,
  "reasoning": "one short sentence in Romanian",
  "behavior_model": "CREATE|INCREASE|REDUCE|MAINTAIN|EVOLVE",
  "directions": ["variant 1", "variant 2", "variant 3"]
}

Behavior model guidance:
- CREATE   = start a new habit/skill/project ("Vreau să învăț Go")
- INCREASE = raise an existing metric ("Vreau să alerg mai mult")
- REDUCE   = lower a metric ("Vreau să slăbesc 10kg")
- MAINTAIN = keep a level ("Vreau să dorm 8h pe noapte")
- EVOLVE   = pivot/transform ("Vreau să schimb cariera spre AI")

Directions: write 2-3 SHORT, CONCRETE reformulations of the goal in Romanian,
each with a measurable target and deadline. If the goal is already specific,
return 1 direction that confirms it. Max 12 words per direction.`

	prompt := fmt.Sprintf("Goal: %q\nDescription: %q\nAnalyze:", title, description)

	text, err := c.complete(ctx, system, prompt, maxTokensShort)
	if err != nil {
		return SuggestionResult{}
	}

	var result SuggestionResult
	if jsonErr := json.Unmarshal([]byte(text), &result); jsonErr != nil {
		return SuggestionResult{}
	}

	// Validate category; drop the whole suggestion if unrecognised.
	validCategories := map[string]bool{
		"HEALTH": true, "CAREER": true, "FINANCE": true,
		"RELATIONSHIPS": true, "LEARNING": true, "CREATIVITY": true, "OTHER": true,
	}
	if !validCategories[result.Category] {
		return SuggestionResult{}
	}
	// Drop BM if Claude returned something off-contract; the caller will fall back.
	if !validBehaviorModels[result.BehaviorModel] {
		result.BehaviorModel = ""
	}
	// Guardrail: cap directions at 3 to keep the UI tidy.
	if len(result.Directions) > 3 {
		result.Directions = result.Directions[:3]
	}
	return result
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// parseLines splits multi-line text into at most maxCount non-empty lines.
func parseLines(text string, maxCount int) []string {
	var lines []string
	buf := make([]byte, 0, len(text))
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			line := string(bytes.TrimSpace(buf))
			if line != "" {
				// Remove leading "- " or "• " markers
				if len(line) > 2 && (line[:2] == "- " || line[:2] == "• ") {
					line = line[2:]
				}
				lines = append(lines, line)
			}
			buf = buf[:0]
		} else {
			buf = append(buf, text[i])
		}
	}
	if line := string(bytes.TrimSpace(buf)); line != "" {
		if len(line) > 2 && (line[:2] == "- " || line[:2] == "• ") {
			line = line[2:]
		}
		lines = append(lines, line)
	}
	if len(lines) > maxCount {
		lines = lines[:maxCount]
	}
	return lines
}
