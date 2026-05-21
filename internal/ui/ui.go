// Package ui provides terminal output helpers for devup.
// It respects the NO_COLOR environment variable (https://no-color.org)
// and wires structured logging to --verbose.
package ui

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

// ANSI color codes.
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

// UI handles all terminal output for devup.
type UI struct {
	out     io.Writer
	noColor bool
}

// New creates a UI that writes to stdout.
// It automatically respects the NO_COLOR env var.
func New(noColor bool) *UI {
	return &UI{
		out:     os.Stdout,
		noColor: noColor || os.Getenv("NO_COLOR") != "",
	}
}

func (u *UI) color(code, text string) string {
	if u.noColor {
		return text
	}
	return code + text + colorReset
}

// Step prints a neutral in-progress step message.
func (u *UI) Step(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(u.out, "  %s %s\n", u.color(colorCyan, "»"), msg)
}

// Success prints a success message.
func (u *UI) Success(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(u.out, "  %s %s\n", u.color(colorGreen, "✓"), msg)
}

// Warn prints a warning message.
func (u *UI) Warn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(u.out, "  %s %s\n", u.color(colorYellow, "!"), msg)
}

// Error prints an error message.
func (u *UI) Error(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(u.out, "  %s %s\n", u.color(colorRed, "✗"), msg)
}

// Header prints a bold section header.
func (u *UI) Header(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(u.out, "\n%s\n", u.color(colorBold, msg))
}

// DoctorRow prints a single doctor check row with aligned columns.
func (u *UI) DoctorRow(ok bool, label, detail string) {
	var icon string
	if ok {
		icon = u.color(colorGreen, "✓")
	} else {
		icon = u.color(colorRed, "✗")
	}
	fmt.Fprintf(u.out, "  %s  %-14s %s\n", icon, label, detail)
}

// DoctorWarnRow prints a doctor check row with a warning indicator.
func (u *UI) DoctorWarnRow(label, detail string) {
	icon := u.color(colorYellow, "!")
	fmt.Fprintf(u.out, "  %s  %-14s %s\n", icon, label, detail)
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
