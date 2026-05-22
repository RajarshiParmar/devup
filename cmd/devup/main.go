package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/RajarshiParmar/devup/internal/brew"
	"github.com/RajarshiParmar/devup/internal/docker"
	devmake "github.com/RajarshiParmar/devup/internal/make"
	"github.com/RajarshiParmar/devup/internal/orbstack"
	"github.com/RajarshiParmar/devup/internal/runner"
	"github.com/RajarshiParmar/devup/internal/ui"
)

// Build-time variables injected via -ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

// globalFlags holds values bound to persistent flags / viper config.
type globalFlags struct {
	dir     string
	timeout int
	upgrade bool
	dryRun  bool
	noColor bool
	verbose bool
}

func newRootCmd() *cobra.Command {
	f := &globalFlags{}

	root := &cobra.Command{
		Use:   "devup",
		Short: "Spin up your local development environment",
		Long: `devup starts your local development environment by:
  1. (optionally) running brew update && brew upgrade
  2. ensuring the Docker context is set to OrbStack
  3. starting OrbStack if the Docker engine is not running
  4. running make start in the project directory`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// Skip viper setup for version/help.
			if cmd.Name() == "version" {
				return nil
			}
			ui.SetupLogger(f.verbose)
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runStart(cmd.Context(), f)
		},
	}

	// Persistent flags (available to all sub-commands).
	pf := root.PersistentFlags()
	pf.BoolVar(&f.noColor, "no-color", false, "disable colored output (env: NO_COLOR)")
	pf.BoolVar(&f.verbose, "verbose", false, "enable debug logging (env: DEVUP_VERBOSE)")

	// Root-command flags.
	cwd, _ := os.Getwd()
	root.Flags().StringVar(&f.dir, "dir", cwd, "project directory to run make start in (env: DEVUP_DIR)")
	root.Flags().IntVar(&f.timeout, "timeout", 60, "max seconds to wait for Docker engine (env: DEVUP_TIMEOUT)")
	root.Flags().BoolVar(&f.upgrade, "upgrade", false, "run brew update && upgrade before starting (env: DEVUP_UPGRADE)")
	root.Flags().BoolVar(&f.dryRun, "dry-run", false, "print commands without executing them")

	// Bind flags to viper so env vars + config file override defaults.
	_ = viper.BindPFlag("dir", root.Flags().Lookup("dir"))
	_ = viper.BindPFlag("timeout", root.Flags().Lookup("timeout"))
	_ = viper.BindPFlag("upgrade", root.Flags().Lookup("upgrade"))
	_ = viper.BindPFlag("verbose", root.PersistentFlags().Lookup("verbose"))

	viper.SetEnvPrefix("DEVUP")
	viper.AutomaticEnv()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/devup")
	_ = viper.ReadInConfig()

	root.AddCommand(newVersionCmd())
	root.AddCommand(newDoctorCmd(f))
	root.AddCommand(newInitCmd(f))

	// Wire context cancellation to OS signals.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	_ = stop
	root.SetContext(ctx)

	return root
}

// ─── start (root) ────────────────────────────────────────────────────────────

func runStart(ctx context.Context, f *globalFlags) error {
	u := ui.New(f.noColor)
	r := runner.New()

	// Resolve effective values from viper (env > flag > default).
	dir := viper.GetString("dir")
	timeout := time.Duration(viper.GetInt("timeout")) * time.Second
	doUpgrade := viper.GetBool("upgrade")

	u.Header("devup  %s", version)

	if f.dryRun {
		u.Warn("dry-run mode — no commands will be executed")
	}

	totalStart := time.Now()

	// Step 1: Optional brew upgrade.
	if doUpgrade {
		if f.dryRun {
			u.Step("running brew update && brew upgrade (skipped: dry-run)")
		} else {
			_, err := u.SpinStep("Updating Homebrew", func() error {
				return brew.UpdateAndUpgrade(ctx, r)
			})
			if err != nil {
				return err
			}
		}
	}

	// Step 2: Switch Docker context to orbstack.
	if f.dryRun {
		u.Step("checking Docker context (skipped: dry-run)")
	} else {
		_, err := u.SpinStep("Switching Docker context", func() error {
			return docker.SwitchContext(ctx, r, "orbstack")
		})
		if err != nil {
			u.Warn("could not switch Docker context: %v", err)
		}
	}

	// Step 3: Ensure Docker engine is running.
	if !f.dryRun && !docker.IsReady(ctx, r) {
		_, err := u.SpinStep("Starting OrbStack", func() error {
			return orbstack.Start(ctx, r)
		})
		if err != nil {
			return err
		}

		_, err = u.SpinStep(fmt.Sprintf("Waiting for Docker engine (timeout: %s)", timeout), func() error {
			return docker.WaitReady(ctx, r, timeout, func(_ time.Duration) {})
		})
		if err != nil {
			return err
		}
	}
	if !f.dryRun {
		u.Success("Docker engine ready")
	}

	// Step 4: Run make start.
	if f.dryRun {
		u.Step("running make start in %s (skipped: dry-run)", dir)
	} else {
		_, err := u.SpinStep(fmt.Sprintf("Running make start in %s", dir), func() error {
			return devmake.Start(ctx, r, dir)
		})
		if err != nil {
			return err
		}
	}

	// Summary.
	totalElapsed := time.Since(totalStart)
	fmt.Println()
	u.Success("All done in %s", formatDuration(totalElapsed))

	return nil
}

// formatDuration formats a duration for the summary line.
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Second:
		return fmt.Sprintf("%dms", d.Milliseconds())
	case d < time.Minute:
		return fmt.Sprintf("%.1fs", d.Seconds())
	default:
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
}

// ─── version ─────────────────────────────────────────────────────────────────

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("devup %s (commit: %s, built: %s)\n", version, commit, date)
		},
	}
}

// ─── doctor ──────────────────────────────────────────────────────────────────

func newDoctorCmd(f *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check all prerequisites without making changes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDoctor(f)
		},
	}
}

type check struct {
	label  string
	ok     bool
	warn   bool
	detail string
}

func runDoctor(f *globalFlags) error {
	u := ui.New(f.noColor)
	r := runner.New()

	u.Header("devup doctor")

	dir := viper.GetString("dir")

	checks := []check{
		checkBinary(r, "docker"),
		checkBinary(r, "make"),
		checkBinary(r, "brew"),
		checkBinary(r, "orb"),
		checkDockerContext(r),
		checkProjectDir(dir),
	}

	allOk := true
	for _, c := range checks {
		switch {
		case c.warn:
			u.DoctorWarnRow(c.label, c.detail)
		case c.ok:
			u.DoctorRow(true, c.label, c.detail)
		default:
			u.DoctorRow(false, c.label, c.detail)
			if !c.warn {
				allOk = false
			}
		}
	}

	fmt.Println()
	if allOk {
		u.Success("all checks passed")
	} else {
		u.Error("one or more required checks failed")
		return fmt.Errorf("doctor found issues")
	}
	return nil
}

func checkBinary(r runner.Runner, name string) check {
	path, err := r.LookPath(name)
	if err != nil {
		// orb and brew are not hard requirements — warn rather than fail.
		optional := name == "orb" || name == "brew"
		return check{
			label:  name,
			ok:     false,
			warn:   optional,
			detail: "not found in PATH",
		}
	}
	return check{label: name, ok: true, detail: path}
}

func checkDockerContext(r runner.Runner) check {
	ctx := context.Background()
	current, err := docker.CurrentContext(ctx, r)
	if err != nil {
		return check{label: "docker ctx", ok: false, detail: "could not read context"}
	}
	if current != "orbstack" {
		return check{
			label:  "docker ctx",
			ok:     false,
			warn:   true,
			detail: fmt.Sprintf("current context is %q (expected orbstack)", current),
		}
	}
	return check{label: "docker ctx", ok: true, detail: current}
}

func checkProjectDir(dir string) check {
	info, err := os.Stat(dir)
	if err != nil {
		return check{label: "project dir", ok: false, detail: fmt.Sprintf("%q not found", dir)}
	}
	if !info.IsDir() {
		return check{label: "project dir", ok: false, detail: fmt.Sprintf("%q is not a directory", dir)}
	}
	return check{label: "project dir", ok: true, detail: dir}
}
