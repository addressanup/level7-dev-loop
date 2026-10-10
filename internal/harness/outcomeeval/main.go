// Command outcomeeval runs the same frozen tasks through Level 7's crew and a
// plain Codex session, and grades both deterministically. It is a
// development harness: it ships in no package and never runs models in CI.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const usageText = `usage:
  outcomeeval validate [--no-checks]
  outcomeeval run --l7 PATH [--confirm] [--smoke] [--codex PATH] [--out DIR] [--tasks a,b] [--rounds N]
  outcomeeval run --l7 PATH --resume RUN_DIR [--confirm] [--codex PATH]
  outcomeeval report --run RUN_DIR`

func main() {
	if filepath.Base(os.Args[0]) == meterName {
		os.Exit(meterMain(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, arguments []string, stdout, stderr io.Writer) int {
	if len(arguments) == 0 {
		fmt.Fprintln(stderr, usageText)
		return 2
	}
	var err error
	switch arguments[0] {
	case "validate":
		err = validateCommand(ctx, arguments[1:], stdout, stderr)
	case "run":
		err = runCommand(ctx, arguments[1:], stdout, stderr)
	case "report":
		err = reportCommand(arguments[1:], stdout, stderr)
	default:
		fmt.Fprintln(stderr, usageText)
		return 2
	}
	var usage usageError
	switch {
	case errors.As(err, &usage):
		fmt.Fprintf(stderr, "outcomeeval: %s\n%s\n", usage.message, usageText)
		return 2
	case err != nil:
		fmt.Fprintln(stderr, "outcomeeval: FAILED:", err)
		return 1
	}
	return 0
}

type usageError struct{ message string }

func (value usageError) Error() string { return value.message }

func parseFlags(name string, arguments []string, stderr io.Writer, define func(*flag.FlagSet)) error {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(stderr)
	define(flags)
	if err := flags.Parse(arguments); err != nil {
		return usageError{err.Error()}
	}
	if flags.NArg() != 0 {
		return usageError{"unexpected argument " + flags.Arg(0)}
	}
	return nil
}

func validateCommand(ctx context.Context, arguments []string, stdout, stderr io.Writer) error {
	skipChecks := false
	if err := parseFlags("validate", arguments, stderr, func(flags *flag.FlagSet) {
		flags.BoolVar(&skipChecks, "no-checks", false, "skip running the base and reference solutions")
	}); err != nil {
		return err
	}
	loaded, err := loadCorpus(embeddedCorpus)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "outcome-eval: protocol %s %s, corpus sha256:%s, %d tasks\n", loaded.protocol.ID, loaded.protocol.Version, loaded.digest, len(loaded.tasks))
	if skipChecks {
		return nil
	}
	if _, err := exec.LookPath("python3"); err != nil {
		return errors.New("python3 is required to check the truth labels; pass --no-checks to skip them")
	}
	work, err := os.MkdirTemp("", "l7-outcome-eval-validate-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)
	judge := grader{rules: loaded.protocol, sandbox: sandboxAvailable(), timeout: time.Duration(loaded.protocol.Timeouts.Check) * time.Second}
	failures := 0
	for _, value := range loaded.tasks {
		if err := truthCheck(ctx, judge, value, work); err != nil {
			failures++
			fmt.Fprintf(stdout, "FAIL %s: %v\n", value.ID, err)
			continue
		}
		fmt.Fprintf(stdout, "PASS %s: base fails, reference passes within scope\n", value.ID)
	}
	if failures != 0 {
		return fmt.Errorf("%d tasks have wrong truth labels", failures)
	}
	return nil
}

func runCommand(ctx context.Context, arguments []string, stdout, stderr io.Writer) error {
	options, tasks := runOptions{}, ""
	if err := parseFlags("run", arguments, stderr, func(flags *flag.FlagSet) {
		flags.StringVar(&options.l7, "l7", "", "absolute path to the l7 binary under evaluation")
		flags.StringVar(&options.codex, "codex", "", "Codex executable (default: codex on PATH)")
		flags.StringVar(&options.out, "out", filepath.Join(".cache", "outcome-eval"), "directory for run records")
		flags.StringVar(&tasks, "tasks", "", "comma-separated task subset; makes the run exploratory")
		flags.IntVar(&options.rounds, "rounds", 0, "rounds per task (default: the protocol's); other values make the run exploratory")
		flags.BoolVar(&options.smoke, "smoke", false, "use a fake Codex that applies reference solutions; no model calls")
		flags.BoolVar(&options.confirm, "confirm", false, "start a live run, which spends Codex quota")
		flags.StringVar(&options.resume, "resume", "", "continue an interrupted run from its directory")
	}); err != nil {
		return err
	}
	if options.l7 == "" {
		return usageError{"--l7 is required"}
	}
	if tasks != "" {
		options.tasks = strings.Split(tasks, ",")
	}
	return runEvaluation(ctx, options, stdout, stderr)
}

func reportCommand(arguments []string, stdout, stderr io.Writer) error {
	directory := ""
	if err := parseFlags("report", arguments, stderr, func(flags *flag.FlagSet) {
		flags.StringVar(&directory, "run", "", "run directory to report")
	}); err != nil {
		return err
	}
	if directory == "" {
		return usageError{"--run is required"}
	}
	loaded, err := loadCorpus(embeddedCorpus)
	if err != nil {
		return err
	}
	manifest, records, runDirectory, err := loadRun(directory)
	if err != nil {
		return err
	}
	report := buildReport(manifest, loaded, records)
	if err := writeReport(runDirectory, report, records); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "outcome-eval: Phase 4 gate %s; improvement claim %s\noutcome-eval: report %s\n", report.Phase4.Result, report.Improvement.Result, filepath.Join(runDirectory, "report.md"))
	return nil
}
