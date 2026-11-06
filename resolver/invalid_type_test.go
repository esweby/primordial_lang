package resolver

import "testing"

func TestResolver_InvalidIsNotAWritableType(t *testing.T) {
	_, errs := parseAndResolve(t, `fn f(a invalid): int { return 1; }`)
	if !containsError(errs, "unknown type: invalid") {
		t.Errorf("expected unknown type error, got %v", errs)
	}
}
