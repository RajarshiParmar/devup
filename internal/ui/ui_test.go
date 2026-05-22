package ui

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{500 * time.Millisecond, "500ms"},
		{1500 * time.Millisecond, "1.5s"},
		{90 * time.Second, "1m30s"},
	}

	for _, tt := range tests {
		got := formatDuration(tt.d)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestSpinStepSuccess(t *testing.T) {
	u := New(true) // no-color for predictable output
	var buf bytes.Buffer
	u.out = &buf

	dur, err := u.SpinStep("Testing step", func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if dur < 100*time.Millisecond {
		t.Errorf("expected duration >= 100ms, got %v", dur)
	}

	output := buf.String()
	if !strings.Contains(output, "✓") {
		t.Errorf("expected success icon in output, got: %q", output)
	}
	if !strings.Contains(output, "Testing step") {
		t.Errorf("expected label in output, got: %q", output)
	}
}

func TestSpinStepError(t *testing.T) {
	u := New(true) // no-color
	var buf bytes.Buffer
	u.out = &buf

	testErr := errors.New("something went wrong")
	_, err := u.SpinStep("Failing step", func() error {
		return testErr
	})

	if !errors.Is(err, testErr) {
		t.Fatalf("expected testErr, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "✗") {
		t.Errorf("expected error icon in output, got: %q", output)
	}
	if !strings.Contains(output, "Failing step") {
		t.Errorf("expected label in output, got: %q", output)
	}
}

func TestOutputMethods(t *testing.T) {
	u := New(true)
	var buf bytes.Buffer
	u.out = &buf

	u.Step("step %s", "msg")
	u.Success("ok %s", "msg")
	u.Warn("warn %s", "msg")
	u.Error("err %s", "msg")
	u.Header("header %s", "msg")

	output := buf.String()
	if !strings.Contains(output, "» step msg") {
		t.Errorf("Step output missing, got: %q", output)
	}
	if !strings.Contains(output, "✓ ok msg") {
		t.Errorf("Success output missing, got: %q", output)
	}
	if !strings.Contains(output, "! warn msg") {
		t.Errorf("Warn output missing, got: %q", output)
	}
	if !strings.Contains(output, "✗ err msg") {
		t.Errorf("Error output missing, got: %q", output)
	}
	if !strings.Contains(output, "header msg") {
		t.Errorf("Header output missing, got: %q", output)
	}
}
