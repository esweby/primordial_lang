package semantic

import (
	"strings"
	"testing"
)

func TestAnalyzeBasicMaps(t *testing.T) {
	_, errors := analyzeProgram(t, `
		map[int32]string{ 10: "ten", }
		map[string]string{ "ten": "ten", "twenty": "twenty one" }
	`)
	if len(errors) != 0 {
		t.Fatalf("expected valid maps, got %v", errors)
	}
}

func TestAnalyzeMapTypeMismatch(t *testing.T) {
    _, errors := analyzeProgram(t, `
        map[int32]string{ "wrong": "ten" }
    `)
    if len(errors) == 0 {
        t.Fatal("expected type mismatch error for key")
    }
    if !strings.Contains(errors[0].Error(), "key") {
        t.Fatalf("expected key error, got: %v", errors[0])
    }
}
