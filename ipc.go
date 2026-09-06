package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// ipcArgs is the parsed form of a `tahubcli ipc <método> [key=value...]
// --json` invocation — the same scriptable-data-source convention as
// dcal/djobs/tabelharadar (github.com/TAbelhaDev/tabelhascaff/ipc). Copied
// locally rather than imported: that package isn't committed/tagged yet, so
// a `replace` would break `go build`/`go install` outside this machine.
// Swap for the real import once tabelhascaff/ipc ships a tagged version.
type ipcArgs struct {
	Method  string
	Filters map[string]string
	JSON    bool
}

func parseIPCArgs(args []string) (*ipcArgs, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("método ausente")
	}
	out := &ipcArgs{Method: args[0], Filters: map[string]string{}}
	for _, arg := range args[1:] {
		if arg == "--json" {
			out.JSON = true
			continue
		}
		if k, v, ok := strings.Cut(arg, "="); ok {
			out.Filters[k] = v
			continue
		}
		return nil, fmt.Errorf("argumento inválido: %q (esperado key=value ou --json)", arg)
	}
	if !out.JSON {
		return nil, fmt.Errorf("apenas saída --json é suportada por enquanto")
	}
	return out, nil
}

func writeJSON(v any) int {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintln(os.Stderr, "erro ao serializar json:", err)
		return 1
	}
	return 0
}

// runIPC implements `tahubcli ipc <método> [key=value...] --json`.
func runIPC(args []string) int {
	parsed, err := parseIPCArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "uso: tahubcli ipc <método> [key=value...] --json")
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	switch parsed.Method {
	case "suggest":
		return ipcSuggest(parsed.Filters)
	default:
		fmt.Fprintf(os.Stderr, "método desconhecido: %q\n", parsed.Method)
		return 1
	}
}

// ipcSuggest reads the projects/postings payload from stdin and prints the
// generated post suggestions. Filters: n= overrides the market-summary top-N
// (default topNPostings).
func ipcSuggest(filters map[string]string) int {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro ao ler stdin:", err)
		return 1
	}

	input, err := parseSuggestInput(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro ao interpretar JSON de entrada:", err)
		return 1
	}

	topN := topNPostings
	if raw, ok := filters["n"]; ok {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			fmt.Fprintln(os.Stderr, "filtro n= inválido, esperado inteiro positivo")
			return 1
		}
		topN = n
	}

	return writeJSON(buildSuggestions(input, topN))
}
