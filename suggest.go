package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// topNPostings is the default number of postings summarized by the market
// suggestion, overridable via the `n=` IPC filter.
const topNPostings = 5

// suggestInput is the tolerant wire format for `tahubcli ipc suggest
// --json`: either a bare array of projects, or an object carrying
// `projects` and (optionally) `postings`. Field sets on both sides are
// still moving in tabelhaglue/tabelhavagas, so unknown fields are ignored
// rather than rejected.
type suggestInput struct {
	Projects []project `json:"projects"`
	Postings []posting `json:"postings"`
}

// project mirrors the fields tabelharadar's `projects.list`/`projects.next`
// IPC output already carries (see tabelharadar/ipc.go's projectJSON) that
// this command actually uses.
type project struct {
	Name          string   `json:"name"`
	DirtyCount    int      `json:"dirty_count"`
	LastCommitMsg string   `json:"last_commit_msg"`
	MemoryNotes   []string `json:"memory_notes"`
	NextSteps     string   `json:"next_steps"`
}

// posting mirrors tabelhavagas's `postings.top`/`postings.list` IPC output.
type posting struct {
	Title    string   `json:"title"`
	Company  string   `json:"company"`
	Location string   `json:"location"`
	Score    int      `json:"score"`
	Tags     []string `json:"tags"`
	URL      string   `json:"url"`
	Deadline string   `json:"deadline"`
}

// suggestion is the post-suggestion contract new-post.mjs (tabelhahub)
// consumes: Slug carries no date prefix (new-post.mjs prefixes the file
// name itself) and Placeholder is markdown body only, no frontmatter.
type suggestion struct {
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Summary     string   `json:"summary"`
	Tags        []string `json:"tags"`
	Placeholder string   `json:"placeholder"`
}

// parseSuggestInput accepts either a bare `[...]` project array or a
// `{"projects": [...], "postings": [...]}` object.
func parseSuggestInput(data []byte) (suggestInput, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return suggestInput{}, fmt.Errorf("nenhum JSON recebido no stdin")
	}
	switch data[0] {
	case '[':
		var projects []project
		if err := json.Unmarshal(data, &projects); err != nil {
			return suggestInput{}, err
		}
		return suggestInput{Projects: projects}, nil
	case '{':
		var in suggestInput
		if err := json.Unmarshal(data, &in); err != nil {
			return suggestInput{}, err
		}
		return in, nil
	default:
		return suggestInput{}, fmt.Errorf("esperado array de projetos ou objeto {projects, postings}")
	}
}

func buildSuggestions(input suggestInput, topN int) []suggestion {
	out := make([]suggestion, 0, len(input.Projects)+1)
	for _, p := range input.Projects {
		if !isActiveProject(p) {
			continue
		}
		out = append(out, projectSuggestion(p))
	}
	if len(input.Postings) > 0 {
		out = append(out, marketSuggestion(input.Postings, topN))
	}
	return out
}

func isActiveProject(p project) bool {
	if p.Name == "" {
		return false
	}
	return p.DirtyCount > 0 || p.LastCommitMsg != "" || len(p.MemoryNotes) > 0
}

var conventionalPrefixRe = regexp.MustCompile(`^([a-z]+)(\([^)]*\))?!?:\s*`)

// commitTypeAndSubject splits a conventional-commit message's first line
// into its type ("feat", "fix", ...; empty if there's no recognized
// prefix) and the subject with that prefix stripped.
func commitTypeAndSubject(msg string) (string, string) {
	line := strings.SplitN(msg, "\n", 2)[0]
	line = strings.TrimSpace(line)
	m := conventionalPrefixRe.FindStringSubmatch(line)
	if m == nil {
		return "", line
	}
	return m[1], strings.TrimSpace(line[len(m[0]):])
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	return strings.ToUpper(string(r[0])) + string(r[1:])
}

func projectTitle(name, commitType, subject string, dirty int) string {
	switch {
	case subject == "" && dirty > 0:
		return fmt.Sprintf("%s: o que tá em andamento", name)
	case commitType == "feat":
		return fmt.Sprintf("Novo no %s: %s", name, capitalize(subject))
	case commitType == "fix":
		return fmt.Sprintf("%s: consertando %s", name, subject)
	default:
		return fmt.Sprintf("%s: %s", name, capitalize(subject))
	}
}

func projectSuggestion(p project) suggestion {
	commitType, subject := commitTypeAndSubject(p.LastCommitMsg)

	title := projectTitle(p.Name, commitType, subject, p.DirtyCount)

	slugSeed := p.Name
	if subject != "" {
		slugSeed += "-" + subject
	}
	slug := slugify(slugSeed)

	var summary strings.Builder
	if subject != "" {
		fmt.Fprintf(&summary, "Último commit: %q.", subject)
	} else {
		summary.WriteString("Ainda sem commits com mensagem descritiva.")
	}
	if p.DirtyCount > 0 {
		fmt.Fprintf(&summary, " %d arquivo(s) ainda em andamento.", p.DirtyCount)
	}
	if len(p.MemoryNotes) > 0 {
		fmt.Fprintf(&summary, " Nota: %s", truncate(p.MemoryNotes[0], 120))
	}

	tags := []string{slugify(p.Name)}
	if commitType == "feat" {
		tags = append(tags, "feature")
	}

	placeholder := projectPlaceholder(p, commitType, subject)

	return suggestion{
		Title:       title,
		Slug:        slug,
		Summary:     truncate(summary.String(), 200),
		Tags:        tags,
		Placeholder: placeholder,
	}
}

const toneGuideBlock = "<!-- GUIA DE TOM (apagar este bloco ao publicar o post)\n" +
	"- Escreva em primeira pessoa, tom direto e conversacional.\n" +
	"- Prefira frases curtas; evite jargão sem explicar.\n" +
	"- Foque no \"por quê\", não só no \"o quê\".\n" +
	"-->"

func projectPlaceholder(p project, commitType, subject string) string {
	lines := []string{
		toneGuideBlock,
		fmt.Sprintf("<!-- contexto coletado pelo tahubcli: último commit %q, %d arquivo(s) sujo(s) -->", p.LastCommitMsg, p.DirtyCount),
		"",
		fmt.Sprintf("Uma frase sobre o que mudou no %s e por que isso importa.", p.Name),
		"",
		"## O que mudou",
		"",
	}
	if subject != "" {
		lines = append(lines, fmt.Sprintf("- %s", capitalize(subject)))
	} else {
		lines = append(lines, "- ")
	}
	lines = append(lines,
		"",
		"## Por que",
		"",
		"",
	)
	if len(p.MemoryNotes) > 0 || p.NextSteps != "" {
		lines = append(lines, "## Anotações do projeto", "")
		for _, note := range p.MemoryNotes {
			lines = append(lines, fmt.Sprintf("- %s", note))
		}
		if p.NextSteps != "" {
			lines = append(lines, fmt.Sprintf("- próximos passos: %s", strings.SplitN(p.NextSteps, "\n", 2)[0]))
		}
		lines = append(lines, "")
	}
	lines = append(lines, "## Limitações", "")

	return strings.Join(lines, "\n")
}

func marketSuggestion(postings []posting, topN int) suggestion {
	filtered := make([]posting, 0, len(postings))
	for _, po := range postings {
		if po.Title == "" {
			continue
		}
		filtered = append(filtered, po)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		return filtered[i].Score > filtered[j].Score
	})
	if len(filtered) > topN {
		filtered = filtered[:topN]
	}

	var highlights []string
	for _, po := range filtered {
		if po.Company != "" {
			highlights = append(highlights, fmt.Sprintf("%s (%s)", po.Title, po.Company))
		} else {
			highlights = append(highlights, po.Title)
		}
	}
	summary := fmt.Sprintf("Top %d vagas coletadas pelo tabelhavagas: %s.", len(filtered), strings.Join(highlights, ", "))

	lines := []string{
		toneGuideBlock,
		"",
		"Uma frase sobre o panorama de vagas desta rodada.",
		"",
		"## As vagas",
		"",
	}
	for _, po := range filtered {
		entry := fmt.Sprintf("- **%s** - %s", po.Title, po.Company)
		if po.Location != "" {
			entry += " - " + po.Location
		}
		entry += fmt.Sprintf(" (score %d)", po.Score)
		if po.Deadline != "" {
			entry += fmt.Sprintf(" (prazo: %s)", po.Deadline)
		}
		if po.URL != "" {
			entry += fmt.Sprintf(" - %s", po.URL)
		}
		lines = append(lines, entry)
	}
	lines = append(lines,
		"",
		"## O que elas têm em comum",
		"",
		"",
		"## O que isso diz do mercado",
		"",
	)

	return suggestion{
		Title:       fmt.Sprintf("Mercado dev: %d vagas que passaram pelo radar", len(filtered)),
		Slug:        "mercado-vagas-radar",
		Summary:     truncate(summary, 200),
		Tags:        []string{"tabelhavagas", "mercado"},
		Placeholder: strings.Join(lines, "\n"),
	}
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return strings.TrimRight(string(r[:max]), " ") + "…"
}

var accentReplacer = strings.NewReplacer(
	"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ç", "c", "ñ", "n",
)

var slugInvalidRe = regexp.MustCompile(`[^a-z0-9]+`)

// slugify lowercases, strips pt-BR accents and collapses anything that
// isn't [a-z0-9] into a single hyphen, trimming hyphens at both ends and
// capping length at 60 chars without splitting a word.
func slugify(s string) string {
	s = strings.ToLower(s)
	s = accentReplacer.Replace(s)
	s = slugInvalidRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")

	const maxLen = 60
	if len(s) > maxLen {
		s = s[:maxLen]
		s = strings.TrimRight(s, "-")
		if idx := strings.LastIndex(s, "-"); idx > 0 {
			s = s[:idx]
		}
	}
	return s
}
