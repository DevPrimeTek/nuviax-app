// Package handlers — unit tests for the rule-based fallback helpers used when
// the AI client is unavailable. These keep onboarding working end-to-end
// without ANTHROPIC_API_KEY by guaranteeing the /goals/suggest-category
// endpoint always returns a valid category + behavior model + directions.
package handlers

import (
	"reflect"
	"testing"
)

func TestFallbackCategory_KeywordMatching(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Vreau să slăbesc 10 kg", "HEALTH"},
		{"alerg 5km", "HEALTH"},
		{"Vreau să lansez un SaaS afacere", "CAREER"},
		{"Promovare la job nou", "CAREER"},
		{"Vreau să economisesc bani pentru investitie", "FINANCE"},
		{"atinge 5000 MRR", "FINANCE"},
		{"Petrec mai mult timp cu familia", "RELATIONSHIPS"},
		{"Invat limba spaniola", "LEARNING"},
		{"Citesc o carte pe luna", "LEARNING"},
		{"Compun muzica", "CREATIVITY"},
		{"Scriu un roman", "CREATIVITY"},
		{"Text ambiguu fara cuvinte cheie", "OTHER"},
		{"", "OTHER"},
	}
	for _, c := range cases {
		if got := fallbackCategory(c.in); got != c.want {
			t.Errorf("fallbackCategory(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFallbackBehaviorModel_KeywordMatching(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Vreau sa slabesc 10 kg", "REDUCE"},
		{"Renunt la fumat", "REDUCE"},
		{"Mentin greutatea", "MAINTAIN"},
		{"Pastrez ritmul de somn", "MAINTAIN"},
		{"Schimb cariera spre AI", "EVOLVE"},
		{"Transform rutina", "EVOLVE"},
		{"Invat Go", "CREATE"},
		{"Lansez un produs SaaS", "CREATE"},
		{"Alerg mai mult", "INCREASE"}, // default neutral
		{"", "INCREASE"},
	}
	for _, c := range cases {
		if got := fallbackBehaviorModel(c.in); got != c.want {
			t.Errorf("fallbackBehaviorModel(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFallbackDirections_EmptyReturnsNil(t *testing.T) {
	if got := fallbackDirections(""); got != nil {
		t.Fatalf("expected nil for empty title, got %v", got)
	}
	if got := fallbackDirections("   "); got != nil {
		t.Fatalf("expected nil for whitespace title, got %v", got)
	}
}

func TestFallbackDirections_Passthrough(t *testing.T) {
	want := []string{"Vreau să alerg 5km în 30 min"}
	got := fallbackDirections("Vreau să alerg 5km în 30 min")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestContainsAny(t *testing.T) {
	if !containsAny("abc def ghi", "def") {
		t.Fatal("expected true when substring is present")
	}
	if containsAny("abc def ghi", "zzz") {
		t.Fatal("expected false when no needle matches")
	}
	if containsAny("abc", "zzz", "yyy") {
		t.Fatal("expected false when no needle matches any")
	}
	if !containsAny("abc def", "zzz", "def") {
		t.Fatal("expected true when at least one needle matches")
	}
}
