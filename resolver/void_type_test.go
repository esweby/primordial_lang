package resolver

import "testing"

func TestResolver_VoidIsNotAWritableType(t *testing.T) {
	for _, input := range []string{
		`fn g(a void) {}`,
		`fn g(): void {}`,
		`struct P { x: void; }`,
	} {
		_, errs := parseAndResolve(t, input)
		if !containsError(errs, "unknown type: void") {
			t.Errorf("%s: expected unknown type error, got %v", input, errs)
		}
	}
}
