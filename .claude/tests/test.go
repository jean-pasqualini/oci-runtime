package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ANSI colors
const (
	colorGreen  = "\033[0;32m"
	colorRed    = "\033[0;31m"
	colorYellow = "\033[0;33m"
	colorBlue   = "\033[0;34m"
	colorReset  = "\033[0m"
)

type Rule struct {
	Description string
	Question    string
	Expected    string
}

type Result struct {
	Rule   Rule
	Actual string
	Pass   bool
}

func loadRules(path string) ([]Rule, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comment = '#'

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("cannot parse CSV: %w", err)
	}

	var rules []Rule
	for i, record := range records {
		if i == 0 && strings.ToLower(record[0]) == "description" {
			continue // skip header
		}
		if len(record) < 3 {
			continue
		}
		rules = append(rules, Rule{
			Description: strings.TrimSpace(record[0]),
			Question:    strings.TrimSpace(record[1]),
			Expected:    strings.ToLower(strings.TrimSpace(record[2])),
		})
	}
	return rules, nil
}

func buildBatchPrompt(rules []Rule) string {
	var sb strings.Builder
	sb.WriteString("Réponds à ces questions par true ou false uniquement, une réponse par ligne dans l'ordre, aucun autre texte :\n")
	for i, r := range rules {
		fmt.Fprintf(&sb, "%d. %s\n", i+1, r.Question)
	}
	return sb.String()
}

func runClaude(prompt, effort, apiKey string) ([]string, error) {
	args := []string{"-p", prompt, "--effort", effort}

	cmd := exec.Command("claude", args...)
	cmd.Env = append(os.Environ(), "ANTHROPIC_API_KEY="+apiKey)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("claude error: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	return lines, nil
}

func getAPIKey() (string, error) {
	// 1. env var
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return key, nil
	}

	// 2. macOS Keychain
	out, err := exec.Command("security", "find-generic-password", "-s", "Claude Code", "-w").Output()
	if err == nil {
		key := strings.TrimSpace(string(out))
		if key != "" {
			return key, nil
		}
	}

	return "", fmt.Errorf("no API key found — set ANTHROPIC_API_KEY or login with Claude Code")
}

func printResult(r Result, verbose bool) {
	if r.Pass {
		fmt.Printf("%s✓ PASS%s — %s\n", colorGreen, colorReset, r.Rule.Description)
		if verbose {
			fmt.Printf("         expected: %s\n", r.Rule.Expected)
		}
	} else {
		fmt.Printf("%s✗ FAIL%s — %s\n", colorRed, colorReset, r.Rule.Description)
		fmt.Printf("         expected: %s%s%s\n", colorGreen, r.Rule.Expected, colorReset)
		fmt.Printf("         actual:   %s%s%s\n", colorRed, r.Actual, colorReset)
	}
}

func printJSON(results []Result) {
	type jsonResult struct {
		Description string `json:"description"`
		Expected    string `json:"expected"`
		Actual      string `json:"actual"`
		Pass        bool   `json:"pass"`
	}

	var out []jsonResult
	for _, r := range results {
		out = append(out, jsonResult{
			Description: r.Rule.Description,
			Expected:    r.Rule.Expected,
			Actual:      r.Actual,
			Pass:        r.Pass,
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}

func main() {
	csvPath := flag.String("rules", ".claude/tests/rules.csv", "path to CSV rules file")
	effort := flag.String("effort", "low", "claude effort level (low|medium|high|xhigh|max)")
	verbose := flag.Bool("verbose", false, "show expected value on pass")
	jsonOut := flag.Bool("json", false, "output results as JSON")
	flag.Parse()

	// Load rules
	rules, err := loadRules(*csvPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%serror:%s %v\n", colorRed, colorReset, err)
		os.Exit(1)
	}
	if len(rules) == 0 {
		fmt.Fprintf(os.Stderr, "%serror:%s no rules found in %s\n", colorRed, colorReset, *csvPath)
		os.Exit(1)
	}

	// Get API key
	apiKey, err := getAPIKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%serror:%s %v\n", colorRed, colorReset, err)
		os.Exit(1)
	}

	if !*jsonOut {
		fmt.Printf("\n%sRunning %d rules...%s\n\n", colorBlue, len(rules), colorReset)
	}

	// Build and run batch prompt
	prompt := buildBatchPrompt(rules)
	lines, err := runClaude(prompt, *effort, apiKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%serror:%s %v\n", colorRed, colorReset, err)
		os.Exit(1)
	}

	// Evaluate results
	var results []Result
	pass, fail := 0, 0

	for i, rule := range rules {
		actual := ""
		if i < len(lines) {
			actual = strings.ToLower(strings.TrimSpace(lines[i]))
		}
		ok := actual == rule.Expected
		if ok {
			pass++
		} else {
			fail++
		}
		results = append(results, Result{Rule: rule, Actual: actual, Pass: ok})
	}

	// Output
	if *jsonOut {
		printJSON(results)
		if fail > 0 {
			os.Exit(1)
		}
		return
	}

	for _, r := range results {
		printResult(r, *verbose)
	}

	// Summary
	fmt.Printf("\n%sResults:%s %s%d passed%s · %s%d failed%s\n\n",
		colorBlue, colorReset,
		colorGreen, pass, colorReset,
		colorRed, fail, colorReset,
	)

	if fail > 0 {
		os.Exit(1)
	}
}
