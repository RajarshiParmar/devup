package runner

import (
	"context"
	"fmt"
	"sync"
)

// Call records a single invocation of Run or Output on the MockRunner.
type Call struct {
	Name string
	Args []string
}

// MockRunner is a test double for Runner. It records all calls and
// returns pre-configured responses.
type MockRunner struct {
	mu      sync.Mutex
	calls   []Call
	runErr  map[string]error  // key: "name arg0 arg1..."
	outData map[string][]byte // key: "name arg0 arg1..."
	outErr  map[string]error
	paths   map[string]string // LookPath results
}

// NewMock returns a new MockRunner ready for use.
func NewMock() *MockRunner {
	return &MockRunner{
		runErr:  make(map[string]error),
		outData: make(map[string][]byte),
		outErr:  make(map[string]error),
		paths:   make(map[string]string),
	}
}

func key(name string, args []string) string {
	k := name
	for _, a := range args {
		k += " " + a
	}
	return k
}

// SetRunError configures the error returned by Run for the given command.
func (m *MockRunner) SetRunError(err error, name string, args ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runErr[key(name, args)] = err
}

// SetOutput configures the output and error returned by Output for the
// given command.
func (m *MockRunner) SetOutput(out []byte, err error, name string, args ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := key(name, args)
	m.outData[k] = out
	m.outErr[k] = err
}

// SetLookPath configures the path returned by LookPath for the given file.
func (m *MockRunner) SetLookPath(file, path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.paths[file] = path
}

// Calls returns all recorded Run/Output invocations.
func (m *MockRunner) Calls() []Call {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]Call, len(m.calls))
	copy(cp, m.calls)
	return cp
}

// Run implements Runner.
func (m *MockRunner) Run(ctx context.Context, name string, args ...string) error {
	m.mu.Lock()
	m.calls = append(m.calls, Call{Name: name, Args: args})
	err := m.runErr[key(name, args)]
	m.mu.Unlock()
	return err
}

// Output implements Runner.
func (m *MockRunner) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	m.mu.Lock()
	m.calls = append(m.calls, Call{Name: name, Args: args})
	k := key(name, args)
	out := m.outData[k]
	err := m.outErr[k]
	m.mu.Unlock()
	return out, err
}

// LookPath implements Runner.
func (m *MockRunner) LookPath(file string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.paths[file]; ok {
		return p, nil
	}
	return "", fmt.Errorf("%q: executable file not found in $PATH", file)
}
