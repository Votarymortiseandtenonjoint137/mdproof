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
	"github.com/runkids/mdproof/internal/sandbox"
	"github.com/runkids/mdproof/internal/upgrade"
)

var version = "dev"

func main() {
	var (
		reportFmt       string
		dryRun          bool
		showVersion     bool
		timeout         time.Duration
		cliBuild        string
		cliSetup        string
		cliTeardown     string
		cliStepSetup    string
		cliStepTeardown string
		failFast        bool
		outputFile      string
		cliIsolation    string
		verbose         countFlag
	)

	flag.StringVar(&reportFmt, "report", "", "output format: json, junit")
	flag.BoolVar(&dryRun, "dry-run", false, "parse and classify only, don't execute")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.DurationVar(&timeout, "timeout", 0, "per-step timeout (default: 2m, or from mdproof.json)")
	flag.StringVar(&cliBuild, "build", "", "command to run once before all runbooks")
	flag.StringVar(&cliSetup, "setup", "", "command to run before each runbook")
	flag.StringVar(&cliTeardown, "teardown", "", "command to run after each runbook")
	flag.StringVar(&cliStepSetup, "step-setup", "", "command to run before each step")
	flag.StringVar(&cliStepTeardown, "step-teardown", "", "command to run after each step")
	flag.BoolVar(&failFast, "fail-fast", false, "stop after first failed step")
	flag.StringVar(&outputFile, "output", "", "write report to file")
	flag.StringVar(&outputFile, "o", "", "write report to file (shorthand)")
	flag.Var(&verbose, "v", "verbosity level (-v or -v -v)")

	var (
		stepsFlag       string
		fromFlag        int
		updateSnapshots bool
		inlineMode      bool
		showCoverage    bool
		coverageMin     int
		watchMode       bool
		strict          bool
	)
	flag.StringVar(&stepsFlag, "steps", "", "only run specific steps (comma-separated: 1,3,5)")
	flag.IntVar(&fromFlag, "from", 0, "run from step N onwards")
	flag.BoolVar(&updateSnapshots, "update-snapshots", false, "update snapshot files instead of comparing")
	flag.BoolVar(&updateSnapshots, "u", false, "update snapshot files (shorthand)")
	flag.BoolVar(&inlineMode, "inline", false, "parse inline test blocks from any .md file")
	flag.BoolVar(&showCoverage, "coverage", false, "show coverage report (no execution)")
	flag.IntVar(&coverageMin, "coverage-min", 0, "minimum coverage score (exit 1 if below)")
	flag.BoolVar(&watchMode, "watch", false, "watch for file changes and re-run")
	flag.BoolVar(&strict, "strict", true, "container-only execution (use --strict=false to allow local)")
	flag.StringVar(&cliIsolation, "isolation", "", "isolation mode: shared, per-runbook")
	flag.Parse()

	if cliIsolation != "" && cliIsolation != "shared" && cliIsolation != "per-runbook" {
		fmt.Fprintf(os.Stderr, "error: invalid --isolation value %q: must be \"shared\" or \"per-runbook\"\n", cliIsolation)
		os.Exit(1)
	}

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

	if a := flag.Args(); len(a) > 0 && a[0] == "sandbox" {
		configDir := "."
		fileCfg, _ := mdproof.LoadConfig(configDir)
		exitCode, err := sandbox.Run(a[1:], fileCfg, version)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		os.Exit(exitCode)
	}

	if updateSnapshots && dryRun {
		fmt.Fprintln(os.Stderr, "error: --update-snapshots and --dry-run are mutually exclusive")
		os.Exit(1)
	}
	if watchMode && showCoverage {
		fmt.Fprintln(os.Stderr, "error: --watch and --coverage are mutually exclusive")
		os.Exit(1)
	}
	if watchMode && dryRun {
		fmt.Fprintln(os.Stderr, "error: --watch and --dry-run are mutually exclusive")
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

	exitCode := 0

	// Coverage mode: analyze only, no execution.
	if showCoverage {
		var entries []mdproof.CoverageEntry
		for _, file := range files {
			f, ferr := os.Open(file)
			if ferr != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", ferr)
				exitCode = 1
				continue
			}
			rb, perr := mdproof.ParseRunbook(f)
			f.Close()
			if perr != nil {
				fmt.Fprintf(os.Stderr, "error parsing %s: %v\n", file, perr)
				exitCode = 1
				continue
			}
			steps := mdproof.ClassifyAll(rb.Steps)
			result := mdproof.AnalyzeCoverage(steps)
			entries = append(entries, mdproof.CoverageEntry{
				File:   filepath.Base(file),
				Result: result,
			})
		}
		mdproof.WriteCoverageReport(os.Stdout, entries)
		if coverageMin > 0 {
			total := mdproof.CoverageTotalScore(entries)
			if total < coverageMin {
				fmt.Fprintf(os.Stderr, "coverage %d%% below minimum %d%%\n", total, coverageMin)
				os.Exit(1)
			}
		}
		os.Exit(exitCode)
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
	strictExplicit := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "strict" {
			strictExplicit = true
		}
	})
	cfg := mdproof.MergeConfig(fileCfg, cliBuild, cliSetup, cliTeardown, cliStepSetup, cliStepTeardown, timeout, strict, strictExplicit, cliIsolation)

	// Strict mode off or watch mode → allow local execution.
	if !cfg.IsStrict() || watchMode {
		os.Setenv("MDPROOF_ALLOW_EXECUTE", "1")
	}

	// Safety: refuse to execute commands outside a container.
	if !dryRun && !mdproof.IsContainerEnv() {
		fmt.Fprintln(os.Stderr, mdproof.ErrNotInContainer)
		os.Exit(1)
	}

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

	// Watch mode: re-run on file changes.
	if watchMode {
		runWatchMode(files, dryRun, effectiveTimeout, cfg, mdproof.RunOptions{
			Steps:          stepNums,
			From:           fromFlag,
			FailFast:       failFast,
			SnapshotUpdate: updateSnapshots,
			StepSetup:      cliStepSetup,
			StepTeardown:   cliStepTeardown,
		}, reportFmt, int(verbose), inlineMode, target)
		return // unreachable — watch loop exits via Ctrl+C
	}

	reports, errs := runAllAndReport(files, dryRun, effectiveTimeout, cfg, mdproof.RunOptions{
		Steps:          stepNums,
		From:           fromFlag,
		FailFast:       failFast,
		SnapshotUpdate: updateSnapshots,
		StepSetup:      cfg.StepSetup,
		StepTeardown:   cfg.StepTeardown,
	}, reportFmt, int(verbose), inlineMode)
	if errs > 0 {
		exitCode = 1
	}
	for _, r := range reports {
		if r.Summary.Failed > 0 {
			exitCode = 1
			break
		}
	}

	// Write report to file if --output is specified.
	if outputFile != "" && len(reports) > 0 {
		outF, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot write output file: %v\n", err)
			os.Exit(1)
		}
		switch reportFmt {
		case "junit":
			mdproof.WriteJUnitReport(outF, reports)
		default:
			if len(reports) == 1 {
				mdproof.WriteJSONReport(outF, reports[0])
			} else {
				mdproof.WriteJSONReports(outF, reports)
			}
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
		StepSetup:      filter.StepSetup,
		StepTeardown:   filter.StepTeardown,
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

func runWatchMode(files []string, dryRun bool, timeout time.Duration, cfg mdproof.Config, filter mdproof.RunOptions, reportFmt string, verbosity int, inline bool, target string) {
	w := mdproof.NewWatcher(files)
	w.Snapshot()

	fmt.Fprintf(os.Stderr, "\nmdproof %s — watching %d file(s)\n\n", version, len(files))

	// Initial run.
	runAllAndReport(files, dryRun, timeout, cfg, filter, reportFmt, verbosity, inline)

	fmt.Fprintf(os.Stderr, "\nWatching for changes... (Ctrl+C to quit)\n")

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		// Re-scan directory for new files.
		var currentFiles []string
		if inline {
			currentFiles, _ = resolveInlineFiles(target)
		} else {
			currentFiles, _ = mdproof.ResolveFiles(target)
		}
		w.SetFiles(currentFiles)

		changed := w.DetectChanges()
		if len(changed) == 0 {
			continue
		}

		fmt.Fprintf(os.Stderr, "\n--- %d file(s) changed ---\n\n", len(changed))
		runAllAndReport(changed, dryRun, timeout, cfg, filter, reportFmt, verbosity, inline)
		fmt.Fprintf(os.Stderr, "\nWatching for changes... (Ctrl+C to quit)\n")
	}
}

func runAllAndReport(files []string, dryRun bool, timeout time.Duration, cfg mdproof.Config, filter mdproof.RunOptions, reportFmt string, verbosity int, inline bool) ([]mdproof.Report, int) {
	var reports []mdproof.Report
	errs := 0
	for _, file := range files {
		name := filepath.Base(file)

		// Per-runbook isolation: create temp HOME and TMPDIR.
		runCfg := cfg
		if cfg.Isolation == "per-runbook" {
			isoDir, err := os.MkdirTemp("", "mdproof-iso-*")
			if err != nil {
				fmt.Fprintf(os.Stderr, "error creating isolation dir: %v\n", err)
				errs++
				continue
			}
			if err := os.MkdirAll(filepath.Join(isoDir, "tmp"), 0755); err != nil {
				fmt.Fprintf(os.Stderr, "error creating isolation tmpdir: %v\n", err)
				os.RemoveAll(isoDir)
				errs++
				continue
			}
			// Copy env map to avoid mutation across iterations.
			runCfg.Env = make(map[string]string, len(cfg.Env)+2)
			for k, v := range cfg.Env {
				runCfg.Env[k] = v
			}
			runCfg.Env["HOME"] = isoDir
			runCfg.Env["TMPDIR"] = filepath.Join(isoDir, "tmp")
			// Cleanup after this runbook.
			defer func(dir string) { os.RemoveAll(dir) }(isoDir)
		}

		rpt, err := runFile(file, name, dryRun, timeout, runCfg, filter, filter.SnapshotUpdate, inline)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error running %s: %v\n", file, err)
			errs++
			continue
		}
		reports = append(reports, rpt)

		if reportFmt == "json" {
			mdproof.WriteJSONReport(os.Stdout, rpt)
		}
	}

	if reportFmt == "junit" && len(reports) > 0 {
		mdproof.WriteJUnitReport(os.Stdout, reports)
	} else if reportFmt != "json" && len(reports) > 0 {
		if len(reports) > 1 {
			mdproof.WritePlainSummary(os.Stdout, reports, verbosity)
		} else {
			mdproof.WriteSingleReport(os.Stdout, reports[0], verbosity)
		}
	}
	return reports, errs
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
