package engine

import (
	"testing"
)

func TestYqEngine(t *testing.T) {
	engine, err := NewEngine("yq")
	if err != nil {
		t.Skip("yq not installed, skipping test")
	}

	input := `{"name": "test", "value": 42}`
	query := ".name"

	result, err := engine.Execute(query, input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Query failed: %s", result.Error)
	}

	expected := "test"
	if result.Output != expected+"\n" {
		t.Errorf("Expected %q, got %q", expected, result.Output)
	}
}

func TestJqEngine(t *testing.T) {
	engine, err := NewEngine("jq")
	if err != nil {
		t.Skip("jq not installed, skipping test")
	}

	input := `{"name": "test", "value": 42}`
	query := ".name"

	result, err := engine.Execute(query, input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Query failed: %s", result.Error)
	}

	expected := "\"test\""
	if result.Output != expected+"\n" {
		t.Errorf("Expected %q, got %q", expected, result.Output)
	}
}

func TestEngineValidation(t *testing.T) {
	engine, err := NewEngine("yq")
	if err != nil {
		t.Skip("yq not installed, skipping test")
	}

	// Empty query should be valid
	if err := engine.Validate(""); err != nil {
		t.Errorf("Empty query should be valid: %v", err)
	}
}