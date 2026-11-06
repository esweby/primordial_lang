package semantic

import "testing"

func TestAnalyze_AnnotatedFunctionLiterals(t *testing.T) {
	errs := analyze(t, `f: function := fn(x int): int { return x; }; f(1);`)
	if !containsError(errs, "cannot call a value of type function") {
		t.Errorf("expected a function value to be uncallable, got %v", errs)
	}

	errs = analyze(t, `x: int := fn() {};`)
	if !containsError(errs, "declaration type mismatch") {
		t.Errorf("expected a type mismatch, got %v", errs)
	}

	errs = analyze(t, `f: fn(int): int := fn(x int): int { if (x == 0) { return 0; } return f(x - 1); }; f(1);`)
	if len(errs) > 0 {
		t.Errorf("a signature type should stay callable, got %v", errs)
	}

	errs = analyze(t, `f: function := fn(x int): int { return x; };`)
	if len(errs) > 0 {
		t.Errorf("a function value should still be storable, got %v", errs)
	}
}
