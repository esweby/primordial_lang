package semantic

import (
	"strings"
	"testing"
)

func TestAnalyzeMemberPropertyTypes(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		symbol       string
		expectedType string
	}{
		{
			name:         "array length",
			input:        `result := [3]int32{}.length;`,
			symbol:       "result",
			expectedType: "int64",
		},
		{
			name:         "slice length",
			input:        `result := []int32{}.length;`,
			symbol:       "result",
			expectedType: "int64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, analyzer, errors := analyzeCollectionInput(t, tt.input)
			if len(errors) != 0 {
				t.Fatalf("expected member property to analyze, got %v", errors)
			}

			symbol, ok := analyzer.Symbols().Get(tt.symbol)
			if !ok {
				t.Fatalf("expected symbol %q to be registered", tt.symbol)
			}
			if symbol.Type().Name() != tt.expectedType {
				t.Fatalf("expected %s, got %s", tt.expectedType, symbol.Type().Name())
			}
		})
	}
}

func TestAnalyzeCollectionMemberMethods(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		symbol       string
		expectedType string
	}{
		{
			name:  "slice append",
			input: `mut values := []int32{}; values.append(1);`,
		},
		{
			name:         "array toSlice",
			input:        `result := [3]int32{}.toSlice();`,
			symbol:       "result",
			expectedType: "[]int32",
		},
		{
			name:         "returned slice property",
			input:        `result := [3]int32{}.toSlice().length;`,
			symbol:       "result",
			expectedType: "int64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, analyzer, errors := analyzeCollectionInput(t, tt.input)
			if len(errors) != 0 {
				t.Fatalf("expected member method to analyze, got %v", errors)
			}

			if tt.symbol == "" {
				return
			}

			symbol, ok := analyzer.Symbols().Get(tt.symbol)
			if !ok {
				t.Fatalf("expected symbol %q to be registered", tt.symbol)
			}
			if symbol.Type().Name() != tt.expectedType {
				t.Fatalf("expected %s, got %s", tt.expectedType, symbol.Type().Name())
			}
		})
	}
}

func TestRejectInvalidMemberUsage(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedMessage string
	}{
		{
			name:            "member on scalar",
			input:           `1.length;`,
			expectedMessage: "type int64 has no members",
		},
		{
			name:            "unknown slice member",
			input:           `[]int32{}.unknown;`,
			expectedMessage: "type []int32 has no member unknown",
		},
		{
			name:            "bare method",
			input:           `[]int32{}.append;`,
			expectedMessage: "method append must be called",
		},
		{
			name:            "property called as method",
			input:           `[]int32{}.length();`,
			expectedMessage: "property length is not callable",
		},
		{
			name:            "wrong append arity",
			input:           `[]int32{}.append();`,
			expectedMessage: "append expects 1 arguments, got 0",
		},
		{
			name:            "wrong append type",
			input:           `[]int32{}.append(true);`,
			expectedMessage: "argument 0 to append: expected int32, got bool",
		},
		{
			name:            "array has no append",
			input:           `[3]int32{}.append(1);`,
			expectedMessage: "type [3]int32 has no member append",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, errors := analyzeCollectionInput(t, tt.input)
			if len(errors) != 1 {
				t.Fatalf("expected one semantic error, got %v", errors)
			}
			if !strings.Contains(errors[0].Error(), tt.expectedMessage) {
				t.Fatalf("expected error containing %q, got %q", tt.expectedMessage, errors[0])
			}
		})
	}
}
