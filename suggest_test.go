package main

import (
	"strings"
	"testing"
)

func TestBuildSuggestionsSkipsInactiveProjects(t *testing.T) {
	in := suggestInput{Projects: []project{
		{Name: "quiet"},
		{Name: "active", DirtyCount: 2, LastCommitMsg: "fix: bug X"},
	}}
	out := buildSuggestions(in, topNPostings)
	if len(out) != 1 {
		t.Fatalf("expected 1 suggestion, got %d: %+v", len(out), out)
	}
	if out[0].Slug != "active-bug-x" {
		t.Fatalf("slug = %q", out[0].Slug)
	}
	if strings.Contains(out[0].Slug, "2026") {
		t.Fatalf("slug must not carry a date prefix: %q", out[0].Slug)
	}
}

func TestBuildSuggestionsMarketSummary(t *testing.T) {
	in := suggestInput{
		Projects: []project{{Name: "p", LastCommitMsg: "chore: y"}},
		Postings: []posting{
			{Title: "A", Company: "X", Score: 5},
			{Title: "B", Company: "Y", Score: 90},
		},
	}
	out := buildSuggestions(in, topNPostings)
	if len(out) != 2 {
		t.Fatalf("expected project + market suggestion, got %d", len(out))
	}
	market := out[len(out)-1]
	if market.Slug != "mercado-vagas-radar" {
		t.Fatalf("slug = %q", market.Slug)
	}
	if !strings.Contains(market.Summary, "B (Y)") {
		t.Fatalf("expected top-scored posting first in summary, got %q", market.Summary)
	}
	for _, tag := range market.Tags {
		if strings.ContainsAny(tag, " ,[]") {
			t.Fatalf("tag %q not a plain token", tag)
		}
	}
}

func TestSlugifyStripsAccentsAndCapsLength(t *testing.T) {
	got := slugify("Projeção de vagas — análise")
	if strings.ContainsAny(got, "çãáé —") {
		t.Fatalf("slugify left non-ascii/space chars: %q", got)
	}
	long := slugify(strings.Repeat("palavra ", 20))
	if len(long) > 60 {
		t.Fatalf("slug too long: %d", len(long))
	}
	if strings.HasPrefix(long, "-") || strings.HasSuffix(long, "-") {
		t.Fatalf("slug has leading/trailing hyphen: %q", long)
	}
}
