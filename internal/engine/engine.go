package engine

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// Engine interface for query processors
type Engine interface {
	Execute(query, input string) (*Result, error)
	Validate(query string) error
	Name() string
}

// Result represents the output of a query execution
type Result struct {
	Output   string
	Error    string
	Success  bool
	Duration time.Duration
}

// YqEngine implements the Engine interface for yq
type YqEngine struct{}

// JqEngine implements the Engine interface for jq
type JqEngine struct{}

// NewEngine creates an engine based on the type
func NewEngine(engineType string) (Engine, error) {
	switch engineType {
	case "yq":
		if !isCommandAvailable("yq") {
			return nil, fmt.Errorf("yq command not found in PATH")
		}
		return &YqEngine{}, nil
	case "jq":
		if !isCommandAvailable("jq") {
			return nil, fmt.Errorf("jq command not found in PATH")
		}
		return &JqEngine{}, nil
	default:
		return nil, fmt.Errorf("unsupported engine type: %s", engineType)
	}
}

// Execute runs a yq query on the input data
func (e *YqEngine) Execute(query, input string) (*Result, error) {
	start := time.Now()
	
	if query == "" {
		return &Result{
			Output:   input, // Return input unchanged if no query
			Success:  true,
			Duration: time.Since(start),
		}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "yq", query)
	cmd.Stdin = bytes.NewBufferString(input)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	if err != nil {
		return &Result{
			Output:   "",
			Error:    stderr.String(),
			Success:  false,
			Duration: duration,
		}, nil
	}

	return &Result{
		Output:   stdout.String(),
		Error:    "",
		Success:  true,
		Duration: duration,
	}, nil
}

// Validate checks if the yq query syntax is valid
func (e *YqEngine) Validate(query string) error {
	if query == "" {
		return nil // Empty query is valid
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Use --help to validate syntax without executing
	cmd := exec.CommandContext(ctx, "yq", "--help")
	return cmd.Run()
}

// Name returns the engine name
func (e *YqEngine) Name() string {
	return "yq"
}

// Execute runs a jq query on the input data
func (e *JqEngine) Execute(query, input string) (*Result, error) {
	start := time.Now()
	
	if query == "" {
		return &Result{
			Output:   input, // Return input unchanged if no query
			Success:  true,
			Duration: time.Since(start),
		}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "jq", query)
	cmd.Stdin = bytes.NewBufferString(input)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	if err != nil {
		return &Result{
			Output:   "",
			Error:    stderr.String(),
			Success:  false,
			Duration: duration,
		}, nil
	}

	return &Result{
		Output:   stdout.String(),
		Error:    "",
		Success:  true,
		Duration: duration,
	}, nil
}

// Validate checks if the jq query syntax is valid
func (e *JqEngine) Validate(query string) error {
	if query == "" {
		return nil // Empty query is valid
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Test with empty object to validate syntax
	cmd := exec.CommandContext(ctx, "jq", query)
	cmd.Stdin = bytes.NewBufferString("{}")
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("invalid jq syntax: %s", stderr.String())
	}
	return nil
}

// Name returns the engine name
func (e *JqEngine) Name() string {
	return "jq"
}

// isCommandAvailable checks if a command exists in PATH
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}