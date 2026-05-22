// Package ui provides terminal output helpers for devup.
// It respects the NO_COLOR environment variable (https://no-color.org)
// and wires structured logging to --verbose.
package ui

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Spinner animation frames.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// UI handles all terminal output for devup.
type UI struct {
	out     io.Writer
	noColor bool
	mu      sync.Mutex

	// lipgloss styles (initialized in constructor).
	successIcon lipgloss.Style
	errorIcon   lipgloss.Style
	warnIcon    lipgloss.Style
	stepIcon    lipgloss.Style
	headerStyle lipgloss.Style
	dimStyle    lipgloss.Style
}

// New creates a UI that writes to stdout.
// It automatically respects the NO_COLOR env var.
func New(noColor bool) *UI {
	nc := noColor || os.Getenv("NO_COLOR") != ""

	u := &UI{
		out:     os.Stdout,
		noColor: nc,
	}

	if nc {
		u.successIcon = lipgloss.NewStyle()
		u.errorIcon = lipgloss.NewStyle()
		u.warnIcon = lipgloss.NewStyle()
		u.stepIcon = lipgloss.NewStyle()
		u.headerStyle = lipgloss.NewStyle()
		u.dimStyle = lipgloss.NewStyle()
	} else {
		u.successIcon = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green
		u.errorIcon = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))   // red
		u.warnIcon = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))    // yellow
		u.stepIcon = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))    // cyan
		u.headerStyle = lipgloss.NewStyle().Bold(true)
		u.dimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // gray
	}

	return u
}

// Step prints a neutral in-progress step message.
func (u *UI) Step(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(u.out, "  %s %s\n", u.stepIcon.Render("»"), msg)
}

// Success prints a success message.
func (u *UI) Success(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(u.out, "  %s %s\n", u.successIcon.Render("✓"), msg)
}

// Warn prints a warning message.
func (u *UI) Warn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(u.out, "  %s %s\n", u.warnIcon.Render("!"), msg)
}

// Error prints an error message.
func (u *UI) Error(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(u.out, "  %s %s\n", u.errorIcon.Render("✗"), msg)
}

// Header prints a bold section header.
func (u *UI) Header(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(u.out, "\n%s\n", u.headerStyle.Render(msg))
}

// DoctorRow prints a single doctor check row with aligned columns.
func (u *UI) DoctorRow(ok bool, label, detail string) {
	var icon string
	if ok {
		icon = u.successIcon.Render("✓")
	} else {
		icon = u.errorIcon.Render("✗")
	}
	fmt.Fprintf(u.out, "  %s  %-14s %s\n", icon, label, detail)
}

// DoctorWarnRow prints a doctor check row with a warning indicator.
func (u *UI) DoctorWarnRow(label, detail string) {
	icon := u.warnIcon.Render("!")
	fmt.Fprintf(u.out, "  %s  %-14s %s\n", icon, label, detail)
}

// SpinStep runs fn while showing an animated spinner with the given label.
// On completion it prints the result with elapsed time.
// Returns the elapsed duration and any error from fn.
func (u *UI) SpinStep(label string, fn func() error) (time.Duration, error) {
	start := time.Now()

	// Start the spinner in a background goroutine.
	done := make(chan struct{})
	go u.animate(label, done)

	// Run the actual work.
	err := fn()
	close(done)

	elapsed := time.Since(start)
	timing := u.dimStyle.Render(fmt.Sprintf("(%s)", formatDuration(elapsed)))

	u.mu.Lock()
	defer u.mu.Unlock()

	// Clear the spinner line.
	fmt.Fprintf(u.out, "\r\033[K")

	if err != nil {
		fmt.Fprintf(u.out, "  %s %s %s\n", u.errorIcon.Render("✗"), label, timing)
	} else {
		fmt.Fprintf(u.out, "  %s %s %s\n", u.successIcon.Render("✓"), label, timing)
	}

	return elapsed, err
}

// animate writes spinner frames to the terminal until done is closed.
func (u *UI) animate(label string, done <-chan struct{}) {
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	i := 0
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			frame := spinnerFrames[i%len(spinnerFrames)]
			u.mu.Lock()
			fmt.Fprintf(u.out, "\r\033[K  %s %s", u.stepIcon.Render(frame), label)
			u.mu.Unlock()
			i++
		}
	}
}

// formatDuration formats a duration as a human-friendly string.
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

// SetupLogger configures the global slog logger.
// When verbose is true, the level is set to Debug; otherwise Info.
func SetupLogger(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(h))
}
