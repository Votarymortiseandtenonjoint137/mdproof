package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/runkids/mdproof"
	"github.com/runkids/mdproof/internal/upgrade"
)

var version = "dev"

func main() {
	var (
		reportFmt   string
		dryRun      bool
		showVersion bool
		timeout     time.Duration
		cliBuild    string
		cliSetup    string
		cliTeardown string
		failFast    bool
		outputFile  string
		verbose     countFlag
	)

	flag.StringVar(&reportFmt, "report", "", "output format: json")
	flag.BoolVar(&dryRun, "dry-run", false, "parse and classify only, don't execute")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.DurationVar(&timeout, "timeout", 0, "per-step timeout (default: 2m, or from mdproof.json)")
	flag.StringVar(&cliBuild, "build", "", "command to run once before all runbooks")
	flag.StringVar(&cliSetup, "setup", "", "command to run before each runbook")
	flag.StringVar(&cliTeardown, "teardown", "", "command to run after each runbook")
	flag.BoolVar(&failFast, "fail-fast", false, "stop after first failed step")
	flag.StringVar(&outputFile, "output", "", "write JSON report to file")
	flag.StringVar(&outputFile, "o", "", "write JSON report to file (shorthand)")
	flag.Var(&verbose, "v", "verbosity level (-v or -v -v)")

	var (
		stepsFlag       string
		fromFlag        int
		updateSnapshots bool
		inlineMode      bool
	)
	flag.StringVar(&stepsFlag, "steps", "", "only run specific steps (comma-separated: 1,3,5)")
	flag.IntVar(&fromFlag, "from", 0, "run from step N onwards")
	flag.BoolVar(&updateSnapshots, "update-snapshots", false, "update snapshot files instead of comparing")
	flag.BoolVar(&updateSnapshots, "u", false, "update snapshot files (shorthand)")
	flag.BoolVar(&inlineMode, "inline", false, "parse inline test blocks from any .md file")
	flag.Parse()

	if showVersion {
		fmt.Println("mdproof", version)
		os.Exit(0)
	}

	if a := flag.Args(); len(a) > 0 && a[0] == "upgrade" {
		if err := upgrade.Run(version); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if updateSnapshots && dryRun {
		fmt.Fprintln(os.Stderr, "error: --update-snapshots and --dry-run are mutually exclusive")
		os.Exit(1)
	}

	// Parse and validate step filter flags.
	var stepNums []int
	if stepsFlag != "" {
		if fromFlag > 0 {
			fmt.Fprintln(os.Stderr, "error: --steps and --from are mutually exclusive")
			os.Exit(1)
		}
		for _, s := range strings.Split(stepsFlag, ",") {
			s = strings.TrimSpace(s)
			n, err := strconv.Atoi(s)
			if err != nil || n < 1 {
				fmt.Fprintf(os.Stderr, "error: invalid step number %q\n", s)
				os.Exit(1)
			}
			stepNums = append(stepNums, n)
		}
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: mdproof [flags] <file.md|directory>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	target := args[0]
	var files []string
	var err error
	if inlineMode {
		files, err = resolveInlineFiles(target)
	} else {
		files, err = mdproof.ResolveFiles(target)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "no runbook files found")
		os.Exit(1)
	}

	// Safety: refuse to execute commands outside a container.
	if !dryRun && !mdproof.IsContainerEnv() {
		fmt.Fprintln(os.Stderr, mdproof.ErrNotInContainer)
		os.Exit(1)
	}

	// Load config: mdproof.json in target directory, CLI flags override.
	configDir := target
	if info, err := os.Stat(target); err == nil && !info.IsDir() {
		configDir = filepath.Dir(target)
	}
	fileCfg, err := mdproof.LoadConfig(configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}
	cfg := mdproof.MergeConfig(fileCfg, cliBuild, cliSetup, cliTeardown, timeout)

	effectiveTimeout := cfg.TimeoutDuration()
	if effectiveTimeout == 0 {
		effectiveTimeout = mdproof.DefaultStepTimeout
	}

	// Run build hook once before all runbooks.
	if cfg.Build != "" && !dryRun {
		buildResult := mdproof.RunBuildHook(cfg.Build)
		if !buildResult.OK {
			fmt.Fprintf(os.Stderr, "build failed (exit %d), aborting\n", buildResult.ExitCode)
			if buildResult.Output != "" {
				fmt.Fprintln(os.Stderr, buildResult.Output)
			}
			os.Exit(1)
		}
	}

	exitCode := 0
	var reports []mdproof.Report

	for _, file := range files {
		name := filepath.Base(file)

		report, runErr := runFile(file, name, dryRun, effectiveTimeout, cfg, mdproof.RunOptions{
			Steps:    stepNums,
			From:     fromFlag,
			FailFast: failFast,
		}, updateSnapshots, inlineMode)

		if runErr != nil {
			fmt.Fprintf(os.Stderr, "error running %s: %v\n", file, runErr)
			exitCode = 1
			continue
		}

		reports = append(reports, report)

		if reportFmt == "json" {
			if err := mdproof.WriteJSONReport(os.Stdout, report); err != nil {
				fmt.Fprintf(os.Stderr, "error: write JSON: %v\n", err)
			}
		}

		if report.Summary.Failed > 0 {
			exitCode = 1
		}
	}

	// Print summary.
	if reportFmt != "json" {
		verbosity := int(verbose)
		if len(reports) > 1 {
			mdproof.WritePlainSummary(os.Stdout, reports, verbosity)
		} else if len(reports) == 1 {
			mdproof.WriteSingleReport(os.Stdout, reports[0], verbosity)
		}
	}

	// Write JSON report to file if --output is specified.
	if outputFile != "" && len(reports) > 0 {
		outF, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot write output file: %v\n", err)
			os.Exit(1)
		}
		if len(reports) == 1 {
			mdproof.WriteJSONReport(outF, reports[0])
		} else {
			mdproof.WriteJSONReports(outF, reports)
		}
		if err := outF.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "error: close output file: %v\n", err)
		}
	}

	os.Exit(exitCode)
}

// runFile runs a single runbook file with the given options.
func runFile(path, name string, dryRun bool, timeout time.Duration, cfg mdproof.Config, filter mdproof.RunOptions, updateSnapshots bool, inline bool) (mdproof.Report, error) {
	f, err := os.Open(path)
	if err != nil {
		return mdproof.Report{}, err
	}
	defer f.Close()

	return mdproof.Run(f, name, mdproof.RunOptions{
		DryRun:         dryRun,
		Timeout:        timeout,
		Setup:          cfg.Setup,
		Teardown:       cfg.Teardown,
		Steps:          filter.Steps,
		From:           filter.From,
		FailFast:       filter.FailFast,
		Env:            cfg.Env,
		SnapshotUpdate: updateSnapshots,
		RunbookDir:     filepath.Dir(path),
		Inline:         inline,
	})
}

// resolveInlineFiles finds all .md files in a path (for inline mode).
func resolveInlineFiles(target string) ([]string, error) {
	info, err := os.Stat(target)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{target}, nil
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			files = append(files, filepath.Join(target, e.Name()))
		}
	}
	return files, nil
}

// countFlag implements flag.Value for counting repeated -v flags.
type countFlag int

func (c *countFlag) String() string { return strconv.Itoa(int(*c)) }
func (c *countFlag) Set(s string) error {
	if s == "true" || s == "" {
		*c++
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("invalid verbosity %q", s)
	}
	*c = countFlag(n)
	return nil
}
func (c *countFlag) IsBoolFlag() bool { return true }
