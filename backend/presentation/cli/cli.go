package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	appintake "github.com/Tattsum/translate-prompt/backend/application/intake"
	"github.com/Tattsum/translate-prompt/backend/application/optimize"
	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	infraBP "github.com/Tattsum/translate-prompt/backend/infrastructure/bestpractice"
)

// Config holds CLI flag values.
type Config struct {
	InputFile     string
	OutputFile    string
	MaxTokens     int
	TargetProfile string
	Tokenizer     string
	ReportFormat  string
	DryRun        bool
	DeepDive      bool
	Workspace     string
}

// Run executes the CLI with the given arguments.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	cfg := Config{}
	fs := flag.NewFlagSet("translate-prompt", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&cfg.InputFile, "i", "", "input file (stdin if omitted)")
	fs.StringVar(&cfg.OutputFile, "o", "", "output file (stdout if omitted)")
	fs.IntVar(&cfg.MaxTokens, "max-tokens", 0, "token budget limit (required)")
	fs.StringVar(&cfg.TargetProfile, "target-profile", "codex", "target profile")
	fs.StringVar(&cfg.Tokenizer, "tokenizer", "cl100k_base", "tokenizer encoding")
	fs.StringVar(&cfg.ReportFormat, "report", "text", "report format: text|json|none")
	fs.BoolVar(&cfg.DryRun, "dry-run", false, "report only, no output file")
	fs.BoolVar(&cfg.DeepDive, "deep-dive", false, "enable intake deep dive")
	fs.StringVar(&cfg.Workspace, "workspace", "", "workspace path for investigation")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if cfg.MaxTokens <= 0 {
		fmt.Fprintln(stderr, "error: --max-tokens is required and must be > 0")
		return 1
	}

	profile, ok := budget.ParseProfile(cfg.TargetProfile)
	if !ok {
		fmt.Fprintf(stderr, "error: unknown target profile %q\n", cfg.TargetProfile)
		return 1
	}

	raw, err := readInput(cfg.InputFile)
	if err != nil {
		fmt.Fprintf(stderr, "error reading input: %v\n", err)
		return 1
	}

	loader, err := infraBP.NewLoader()
	if err != nil {
		fmt.Fprintf(stderr, "error loading rules: %v\n", err)
		return 1
	}

	optCfg := budget.Config{
		MaxTokens:     cfg.MaxTokens,
		TargetProfile: profile,
		Tokenizer:     cfg.Tokenizer,
		DeepDive:      cfg.DeepDive,
		WorkspacePath: cfg.Workspace,
	}

	promptText := raw
	if cfg.DeepDive {
		intakeUC := appintake.NewUseCase(loader)
		result, err := intakeUC.Analyze(ctx, raw, optCfg)
		if err != nil {
			fmt.Fprintf(stderr, "error analyzing: %v\n", err)
			return 1
		}
		if result.Status == "needs_input" {
			fmt.Fprintln(stderr, "Intake questions (answer via web UI or enrich prompt):")
			for _, q := range result.Questions {
				fmt.Fprintf(stderr, "  [%s] %s\n", q.ID, q.Text)
			}
		} else {
			promptText = result.Prompt
		}
	}

	if cfg.Workspace != "" {
		intakeUC := appintake.NewUseCase(loader)
		inv, err := intakeUC.Investigate(ctx, cfg.Workspace, profile)
		if err != nil {
			fmt.Fprintf(stderr, "warning: workspace investigation failed: %v\n", err)
		} else {
			promptText = appintake.MergeContext(promptText, inv)
			optCfg.VerificationCommands = inv.SuggestedCommands
		}
	}

	optUC, err := optimize.NewUseCase(loader, cfg.Tokenizer)
	if err != nil {
		fmt.Fprintf(stderr, "error creating optimizer: %v\n", err)
		return 1
	}

	result, err := optUC.Optimize(ctx, promptText, optCfg)
	if err != nil {
		fmt.Fprintf(stderr, "error optimizing: %v\n", err)
		return 1
	}

	if err := writeReport(result, cfg.ReportFormat, stdout); err != nil {
		fmt.Fprintf(stderr, "error writing report: %v\n", err)
		return 1
	}

	if cfg.DryRun {
		return 0
	}

	if err := writeOutput(cfg.OutputFile, result.OptimizedPrompt); err != nil {
		fmt.Fprintf(stderr, "error writing output: %v\n", err)
		return 1
	}

	return 0
}

func readInput(path string) (string, error) {
	var r io.Reader = os.Stdin
	if path != "" {
		f, err := os.Open(path)
		if err != nil {
			return "", err
		}
		defer func() { _ = f.Close() }()
		r = f
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func writeOutput(path, content string) error {
	if path == "" {
		_, err := fmt.Fprint(os.Stdout, content)
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func writeReport(result optimize.Result, format string, w io.Writer) error {
	switch strings.ToLower(format) {
	case "none":
		return nil
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(result.Report)
	default:
		r := result.Report
		_, err := fmt.Fprintf(w, "Input: %d tokens → Output: %d tokens (%.1f%% reduction)\nProfile: %s\n",
			r.InputTokens, r.OutputTokens, r.ReductionPercent, r.TargetProfile)
		if err != nil {
			return err
		}
		for _, rule := range r.AppliedRules {
			fmt.Fprintf(w, "  rule %s (%s)\n", rule.ID, rule.SourceURL)
		}
		return nil
	}
}
