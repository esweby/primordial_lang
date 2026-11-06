package evaluator

import "testing"

func TestAssignmentCopies(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"slice append", `a := []int{1}; mut b := a; b.append(2); len(a);`, 1},
		{"the copy changes", `a := []int{1}; mut b := a; b.append(2); len(b);`, 2},
		{"slice element", `a := []int{1}; b := a; b[0] = 5; a[0];`, 1},
		{"array element", `a := [1]int{1}; b := a; b[0] = 5; a[0];`, 1},
		{"struct field", `struct P { pub x: int; } a := P{x: 1}; b := a; b.x = 2; a.x;`, 1},
		{"map value", `a := map[string]int{"k": 1}; b := a; b["k"] = 2; a["k"];`, 1},
		{"map key", `a := map[string]int{"k": 1}; b := a; b["j"] = 2; mut n := 0; for k, v := range a { n = n + 1; } n;`, 1},
		{"nested in a struct", `struct P { pub xs: []int; } a := P{xs: []int{1}}; b := a; b.xs.append(2); len(a.xs);`, 1},
		{"reassignment", `a := []int{1}; mut b := []int{}; b = a; b.append(2); len(a);`, 1},
		{"into a field", `struct P { pub xs: []int; } a := []int{1}; p := P{xs: []int{}}; p.xs = a; p.xs.append(2); len(a);`, 1},
		{"tuple", `fn f(): []int, int { return []int{1}, 2; } (a, b) := f(); mut c := a; c.append(2); len(a);`, 1},
		{"methods still change self", `struct P { pub x: int; impl { fn set(x int) { self.x = x; } } } p := P{x: 1}; p.set(2); p.x;`, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testIntegerObject(t, testEval(t, tt.input), tt.expected)
		})
	}
}
