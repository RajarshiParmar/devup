package make_test

import (
	"context"
	"errors"
	"os"
	"testing"

	devmake "github.com/RajarshiParmar/devup/internal/make"
	"github.com/RajarshiParmar/devup/internal/runner"
)

func TestStart_Success(t *testing.T) {
	m := runner.NewMock()
	m.SetLookPath("make", "/usr/bin/make")

	dir := t.TempDir()

	err := devmake.Start(context.Background(), m, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, c := range m.Calls() {
		if c.Name == "make" && len(c.Args) == 1 && c.Args[0] == "start" {
			found = true
		}
	}
	if !found {
		t.Error("expected make start to be called")
	}
}

func TestStart_MakeNotFound(t *testing.T) {
	m := runner.NewMock()
	// make not in PATH

	err := devmake.Start(context.Background(), m, t.TempDir())
	if err == nil {
		t.Fatal("expected error when make not in PATH")
	}
}

func TestStart_DirNotFound(t *testing.T) {
	m := runner.NewMock()
	m.SetLookPath("make", "/usr/bin/make")

	err := devmake.Start(context.Background(), m, "/nonexistent/path/xyz")
	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}
}

func TestStart_DirIsFile(t *testing.T) {
	m := runner.NewMock()
	m.SetLookPath("make", "/usr/bin/make")

	f, err := os.CreateTemp(t.TempDir(), "not-a-dir")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	err = devmake.Start(context.Background(), m, f.Name())
	if err == nil {
		t.Fatal("expected error when path is a file, not a directory")
	}
}

func TestStart_MakeFails(t *testing.T) {
	m := runner.NewMock()
	m.SetLookPath("make", "/usr/bin/make")
	m.SetRunError(errors.New("exit status 2"), "make", "start")

	err := devmake.Start(context.Background(), m, t.TempDir())
	if err == nil {
		t.Fatal("expected error when make start fails")
	}
}
