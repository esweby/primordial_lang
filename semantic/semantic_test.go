package semantic

import (
	"strings"
	"testing"

	"github.com/esweby/primordial_lang/inference"
	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/parser"
	"github.com/esweby/primordial_lang/resolver"
)

func TestAnalyze_ValidPrograms(t *testing.T) {
	tests := []string{
		``,
		`2 + 2 + 2 + 2`,
		`!true`,
		`-10`,
		`"a" + "b" == "ab"`,
		`fn add(x int32): int32 { return 10 + 1; } add(5);`,
		`if (1 > 2) { 1; 2; 3; };`,
		`brian := if (2 > 1) { 1 + 2; } else { 1 - 0 }`,
		`x := if (true) { return; } else { 1 };`,
		`mut brian := 1; brian = 2`,
		`struct Person { pub name: string; fn new(): Person { return Person{name: "jenny"}; } } p := Person.new(); p.name = "Bob";`,
		`x := map[int32]string{}; x[12] = "one";`,
		`x := map[string]int{}; x["key"] = 42;`,
		`mut i := int32(1); x := map[int32]string{}; x[i] = "one";`,
		`map[int32]string{10: "ten"}; map[string]string{"ten": "ten", "twenty": "twenty one"};`,
		`add := fn() {}`,
		`add: function := fn(x int32, y int32): int32 { return x + y; }`,
		`add := fn(x int32, y int32): int32 { if (x > y) { return y + x; } else { return x; } }`,
		`fn add(x int32, y int32): int32, bool { return x + y, true; }`,
		`fn add(x int32, y int32) { return; }`,
		`fn sign(x int32): int32 { if (x < 0) { return -1; } else if (x > 0) { return 1; } else { return 0; } }`,
		`newAdder := fn(x int64): fn(int64): int64 { return fn(y int64): int64 { return x + y; }; }; addFive := newAdder(5); total := addFive(5) + 1;`,
		`fn apply(f fn(int64): int64, v int64): int64 { return f(v); } double := fn(n int64): int64 { return n * 2; }; x := apply(double, 4);`,
		`f: function := fn(n int32): int32 { return n; };`,
		`fn store(f function) {} store(fn(n int32): int32 { return n; });`,
		`for {}`,
		`for (1 < 2) {}`,
		`for (x := 1; x < 10; x = x + 1) {}`,
		`for a, b := range map[int32]string{} {}`,
		`lbl: for { break lbl; }`,
		`x := lbl: for { break lbl (1); }`,
		`outer: for { for { break outer; } }`,
		`values := [5]int32{1, 2}; empty := [3]int32{}; slice := []int32{1, 2, 3}; first := values[0];`,
		`fn values(): int32, bool { return 10, true; } (number, _) := values(); number;`,
		`fn values(): int32, int32 { return 10, 20; } mut first: int32 := 0; mut second: int32 := 0; (first, second) = values();`,
		`mut values := []int32{}; values.append(1); n := values.length; s := [3]int32{}.toSlice();`,
		`n := len("hello") + len([]int32{});`,
		`struct Address { pub city: string; } struct Person { name: string; pub address: Address; age: int64 = 42; } person := Person{name: "Ada", address: Address{city: "London"}}; city := person.address.city;`,
		`struct Person { name: string; } name := "Ada"; person := Person{name};`,
		`struct Person { name: string; pub age: int32; fn new(name string, age int32): Person { return Person{name, age}; } impl { fn getName(): string { return self.name; } fn hasName(other Person): bool { return self.name == other.name; } fn birthday(): int32 { self.age = self.age + 1; return self.age; } } } person := Person.new("Ada", 41); name := person.getName(); matches := person.hasName(person); age := person.birthday();`,
		`struct Person { pub name: string; impl { fn setName(name string) { self.name = name; } } } person := Person{name: "Ada"}; person.setName("Grace");`,
		`struct Address { pub city: string; impl { fn setCity(city string) { self.city = city; } } } struct Person { pub address: Address; } person := Person{address: Address{city: "London"}}; person.address.setCity("Paris");`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			if errs := analyze(t, input); len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}
		})
	}
}

func TestAnalyze_Errors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Operators
		{`-true`, "operator - requires a number, got bool"},
		{`!5`, "operator ! requires a bool, got int"},
		{`"a" < "b"`, "mismatched types: string < string"},
		{`true == "a"`, "mismatched types: bool == string"},
		{`fn f() {} fn g() {} f == g`, "type fn(): void cannot be compared"},
		{`fn add(x int32, y int32): int32 { return x + true; }`, "mismatched types: int32 + bool"},

		// Declarations and assignment
		{`brian: int32 := true;`, "declaration type mismatch: expected int32, got bool"},
		{`add: int32 := fn() {}`, "declaration type mismatch: expected int32, got fn(): void"},
		{`values: [2]int32 := [3]int32{1};`, "declaration type mismatch: expected [2]int32, got [3]int32"},
		{`brian := 1; brian = 2;`, "cannot assign to immutable variable: brian"},
		{`mut name := "a"; name = true;`, "assignment type mismatch: expected string, got bool"},
		{`fn f(a int32) { a = 1; }`, "cannot assign to non-variable: a"},
		{`fn values(): int32, int32 { return 10, 20; } first: int32 := 0; mut second: int32 := 0; (first, second) = values();`, "cannot assign to immutable variable: first"},
		{`fn values(): string, bool { return "a", true; } mut a := 1; mut b := true; (a, b) = values();`, "tuple assignment value 0: expected int, got string"},
		{`fn one(): int32 { return 1; } (a, b) := one();`, "tuple declaration requires a tuple value, got int32"},
		{`fn two(): int32, int32 { return 1, 2; } (a, b, c) := two();`, "tuple declaration arity mismatch: expected 2 names, got 3"},
		{`struct P { pub n: int32; impl { fn get(): int32 { return self.n; } } } p := P{n: 1}; p.get = "x";`, "cannot assign to method: get"},

		// Functions and returns
		{`fn add(x int32): int32 { return x + 1; } add(true);`, "argument 0: expected int32, got bool"},
		{`fn f(a int32) {} f(1, 2);`, "wrong number of arguments: expected 1, got 2"},
		{`x := 1; x();`, "cannot call non-function int"},
		{`fn add(x int32, y int32): int32, bool { return x + y; }`, "wrong number of return values: expected 2, got 1"},
		{`fn f(): string { return true; }`, "return value 0: expected string, got bool"},
		{`fn f() { return 1; }`, "unexpected return with values in void function"},
		{`brian := if (2 > 1) { return 1 + 2; } else { 1 - 0 }`, "unexpected return with values in void function"},
		{`fn f(): int32 { if (true) { return 1; } }`, "function 'f' declares return types but does not return on all paths"},
		{`f := fn(): int32 {};`, "function literal declares return types but does not return on all paths"},
		{`len(1)`, "argument to len not supported, got int"},
		{`len()`, "len expects 1 argument, got 0"},
		{`x := len;`, "builtin len must be called"},
		{`newAdder := fn(x int64): function { return fn(y int64): int64 { return x + y; }; }; addFive := newAdder(5); addFive(5);`, "cannot call a value of type function"},
		{`fn apply(f fn(int64): int64) {} apply(fn(n int32): int32 { return n; });`, "argument 0: expected fn(int64): int64, got fn(int32): int32"},
		{`fn apply(f fn(int64): int64) {} fn wrap(g function) { apply(g); }`, "argument 0: expected fn(int64): int64, got function"},
		{`f: fn(string) := fn(n int32) {};`, "declaration type mismatch: expected fn(string): void, got fn(int32): void"},
		{`f: fn(string) := fn(s string) {}; f("a", "b");`, "wrong number of arguments: expected 1, got 2"},

		// If and loops
		{`brian := if (2) { 1 + 2 } else { 1 - 0 }`, "if condition must be boolean, got int"},
		{`x := if (true) { 1 };`, "if expression used as value requires an else branch"},
		{`x := if (true) { "a" } else { true };`, "if expression branches have mismatched types: string vs bool"},
		{`x := if (true) { return; } else { return; };`, "if expression branches both return, no value produced"},
		{`x := if (true) { 1; y := 2; } else { 3 };`, "if expression branches must end with an expression"},
		{`for (1) {}`, "while condition must be boolean, got int"},
		{`for (i := 0; i; i = i + 1) {}`, "loop condition must be boolean, got int"},
		{`for x := range 5 {}`, "range can only iterate over an array, slice, or map, got int"},
		{`x := for {};`, "loop used as expression must have a break to exit"},
		{`x := lbl: for { if (true) { break lbl; } break lbl (1); };`, "mixed break statements: some with values and some without"},
		{`x := lbl: for { if (true) { break lbl ("a"); } break lbl (true); };`, "break value types must match: string and bool"},

		// Collections and indexing
		{`map[int32]string{"wrong": "ten"}`, "pair 0: key: expected int32, got string"},
		{`map[int32]string{1: true}`, "pair 0: value: expected string, got bool"},
		{`values := [2]int32{1, 2, 3};`, "array length is 2, but 3 elements were supplied"},
		{`values := [2]int32{1, true};`, "array element 1: expected int32, got bool"},
		{`values := []string{"a", false};`, "slice element 1: expected string, got bool"},
		{`x := 1; x[0];`, "type int cannot be indexed"},
		{`m := map[string]int32{}; v := m[true];`, "index type mismatch: got bool, want string"},

		// Members
		{`1.length;`, "type int has no members"},
		{`[]int32{}.unknown;`, "type []int32 has no member unknown"},
		{`[]int32{}.append;`, "method append must be called"},
		{`[]int32{}.length();`, "property length is not callable"},
		{`[]int32{}.append();`, "append expects 1 arguments, got 0"},
		{`[]int32{}.append(true);`, "argument 0 to append: expected int32, got bool"},
		{`[3]int32{}.append(1);`, "type [3]int32 has no member append"},
		{`items := []int64{}; items.append(1);`, "cannot call mutating method append on immutable variable: items"},

		// Structs
		{`struct Person { name: string; } Person{}`, "missing required field Person.name"},
		{`struct Person { name: string; } Person{name: "Ada", extra: 1}`, "type Person has no field extra"},
		{`struct Person { name: string; } Person{name: "Ada", name: "Grace"}`, "field name supplied more than once"},
		{`struct Person { name: string; } Person{name: true}`, "field Person.name: expected string, got bool"},
		{`struct Example { value; }`, "field Example.value requires a type or default value"},
		{`struct Example { value: int64; value: int64; }`, "field 'value' already declared on Example"},
		{`struct Example { value: string = true; }`, "default value for Example.value: expected string, got bool"},
		{`struct Example { fn make() {} fn make() {} }`, "type function 'make' already declared on Example"},
		{`struct Example { impl { fn run() {} fn run() {} } }`, "method 'run' already declared on Example"},
		{`struct Example { run: int32; impl { fn run() {} } }`, "member 'run' already declared on Example"},
		{`struct Person {} value := Person;`, "type Person cannot be used as a value"},
		{`struct Person { name: string; impl { fn getName(): string { return self.name; } } } Person.getName();`, "type Person has no type function getName"},
		{`struct Person { name: string; } person := Person{name: "Ada"}; person.name;`, "field Person.name is private"},
		{`struct Person { name: string; } person := Person{name: "Ada"}; person.name = "Grace";`, "field Person.name is private"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			errs := analyze(t, test.input)
			if !containsError(errs, test.expected) {
				t.Fatalf("expected error containing %q, got %v", test.expected, errs)
			}
		})
	}
}

func TestAnalyze_ErrorsIncludePosition(t *testing.T) {
	errs := analyze(t, "x := 1;\nx = 2;")
	if len(errs) != 1 || !strings.HasPrefix(errs[0], "2:1: ") {
		t.Fatalf("expected a single error at 2:1, got %v", errs)
	}
}

func TestAnalyze_RequiresResolvedProgram(t *testing.T) {
	program := parser.New(lexer.New(`x := 1;`)).ParseProgram()
	if errs := New(program).Analyze(); !containsError(errs, "analysis requires a resolved and inferred program") {
		t.Fatalf("expected an error for an unresolved program, got %v", errs)
	}
}

// analyze runs every phase up to the analyzer and returns the analyzer's
// errors. Earlier phases must succeed.
func analyze(t *testing.T, input string) []string {
	t.Helper()
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	if _, errs := resolver.New(program).Resolve(); len(errs) > 0 {
		t.Fatalf("resolver errors: %v", errs)
	}
	if errs := inference.New(program).Infer(); len(errs) > 0 {
		t.Fatalf("inference errors: %v", errs)
	}
	return New(program).Analyze()
}

func containsError(errs []string, expected string) bool {
	for _, err := range errs {
		if strings.Contains(err, expected) {
			return true
		}
	}
	return false
}
