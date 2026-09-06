package main

import "testing"

func TestUnmarshalFilterArrayEmpty(t *testing.T) {
	out, err := unmarshalFilterArray[project]("")
	if err != nil {
		t.Fatalf("unmarshalFilterArray: %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil for empty filter, got %+v", out)
	}
}

func TestUnmarshalFilterArrayProjects(t *testing.T) {
	out, err := unmarshalFilterArray[project](`[{"name":"taradar","dirty_count":1,"last_commit_msg":"feat: x"}]`)
	if err != nil {
		t.Fatalf("unmarshalFilterArray: %v", err)
	}
	if len(out) != 1 || out[0].Name != "taradar" {
		t.Fatalf("got %+v", out)
	}
}

func TestUnmarshalFilterArrayInvalid(t *testing.T) {
	if _, err := unmarshalFilterArray[project]("not json"); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestIpcSuggestFiltersRoundTrip(t *testing.T) {
	filters := map[string]string{
		"projects": `[{"name":"taradar","dirty_count":1,"last_commit_msg":"feat: x"}]`,
		"vagas":    `[{"title":"Go Dev","company":"Acme","score":10}]`,
		"docs":     `[{"name":"taradar","description":"..."}]`,
		"group":    "tabeladev",
	}
	projects, err := unmarshalFilterArray[project](filters["projects"])
	if err != nil {
		t.Fatalf("projects: %v", err)
	}
	postings, err := unmarshalFilterArray[posting](filters["vagas"])
	if err != nil {
		t.Fatalf("vagas: %v", err)
	}
	out := buildSuggestions(suggestInput{Projects: projects, Postings: postings}, topNPostings)
	if len(out) != 2 {
		t.Fatalf("expected project + market suggestion, got %d: %+v", len(out), out)
	}
}
