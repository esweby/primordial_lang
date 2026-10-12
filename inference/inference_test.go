package inference

import (
	"strings"
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/parser"
	"github.com/esweby/primordial_lang/resolver"
	"github.com/esweby/primordial_lang/types"
)

func TestInfer_DeclarationTypes(t *testing.T) {
	tests := []struct {
		input string
		name  string
		want  string
	}{
		{`x := 0;`, "x", "int"},
		{`x := -5 + 2 * 3;`, "x", "int"},
		{`x: int32 := 10;`, "x", "int32"},
		{`y: int32 := 10; w := y + 1;`, "w", "int32"},
		{`y: int32 := 10; w := 1 + y;`, "w", "int32"},
		{`y: int32 := 10; a := int32(0) + y;`, "a", "int32"},
		{`x := 1 < 2;`, "x", "bool"},
		{`x := "a" + "b";`, "x", "string"},
		{`x := len([]int32{});`, "x", "int"},
		{`x := uint8(200);`, "x", "uint8"},
		{`x: uint := 1;`, "x", "uint"},
		{`s := []int32{1}; x := s[0];`, "x", "int32"},
		{`m := map[string]uint8{"a": 1}; x := m["a"];`, "x", "uint8"},
		{`x := if (true) { 1 } else { 2 };`, "x", "int"},
		{`y: int32 := 1; x := if (true) { y } else { 2 };`, "x", "int32"},
		{`y: int32 := 1; x := if (true) { 2 } else if (false) { 3 } else { y };`, "x", "int32"},
		{`x := lbl: for { break lbl (1); };`, "x", "int"},
		{`y: int8 := 1; x := lbl: for { if (true) { break lbl (y); } break lbl (2); };`, "x", "int8"},
		{`fn pair(): int32, bool { return 1, true; } (a, b) := pair();`, "a", "int32"},
		{`fn pair(): int32, bool { return 1, true; } (a, b) := pair();`, "b", "bool"},
		{`fn add(a int8, b int8): int8 { return a + b; } x := add(1, 2);`, "x", "int8"},
		{`f := fn(n int16): int16 { return f(n - 1); }; x := f(3);`, "x", "int16"},
		{`struct P { pub age: int32; } p := P{age: 1}; x := p.age;`, "x", "int32"},
		{`struct P { count = 5; } p := P{}; x := p.count;`, "x", "int"},
		{`struct P { pub age: int32; impl { fn older(): int32 { return self.age + 1; } } } p := P{age: 1}; x := p.older();`, "x", "int32"},
		{`struct P { fn make(): int64 { return 1; } } x := P.make();`, "x", "int64"},
		{`s := []int32{1}; x := s.length;`, "x", "int"},
		{`x: int32 := 1; y := int64(x);`, "y", "int64"},
		{`x := [5]int32{1, 2};`, "x", "[5]int32"},
		{`x := [3]int32{};`, "x", "[3]int32"},
		{`x := []int32{1, 2, 3};`, "x", "[]int32"},
		{`x := [3]int32{}.toSlice();`, "x", "[]int32"},
		{`newAdder := fn(x int64): fn(int64): int64 { return fn(y int64): int64 { return x + y; }; }; addFive := newAdder(5); x := addFive(5);`, "x", "int64"},
		{`fn apply(f fn(int8): int8, v int8): int8 { return f(v); } x := apply(fn(n int8): int8 { return n; }, 3);`, "x", "int8"},
		{`f: fn(): (int32, bool) := fn(): int32, bool { return 1, true; }; (x, ok) := f();`, "x", "int32"},
		{`x := [3]int32{}.toSlice().length;`, "x", "int"},
		{`struct Address { pub city: string; } struct Person { pub address: Address; } p := Person{address: Address{city: "London"}}; x := p.address.city;`, "x", "string"},
		{`struct P { name: string; impl { fn getName(): string { return self.name; } } } p := P{name: "Ada"}; x := p.getName();`, "x", "string"},
		{`struct P { name: string; impl { fn same(other P): bool { return self.name == other.name; } } } p := P{name: "Ada"}; x := p.same(p);`, "x", "bool"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			scope := mustInfer(t, test.input)
			if got := typeOfSymbol(t, scope, test.name); got != test.want {
				t.Fatalf("%s has type %s, want %s", test.name, got, test.want)
			}
		})
	}
}

func TestInfer_IntegerErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`x := 0; y: int32 := 10; z := x + y;`, "mismatched types int and int32"},
		{`x: int8 := 300;`, "integer constant 300 overflows int8"},
		{`x: uint8 := -1;`, "integer constant -1 overflows uint8"},
		{`x: int8 := 100 + 100;`, "integer constant 200 overflows int8"},
		{`x := 1 / 0;`, "division by zero"},
		{`y: int32 := 4; z := y / 0;`, "division by zero"},
		{`x: uint32 := 1; y := -x;`, "cannot negate unsigned integer uint32"},
		{`x: string := 1;`, "cannot use untyped integer constant as string"},
		{`x := "a" + 1;`, "cannot use untyped integer constant as string"},
		{`x: int64 := 1; y: int32 := x;`, "cannot use int64 value as int32"},
		{`x := 5; y: int64 := x;`, "cannot use int value as int64"},
		{`mut x: int8 := 0; y: int16 := 1; x = y;`, "cannot use int16 value as int8"},
		{`mut x: int8 := 0; x = 128;`, "integer constant 128 overflows int8"},
		{`fn f(a int32) {} x := 1; f(x);`, "cannot use int value as int32"},
		{`fn f(a uint8) {} f(256);`, "integer constant 256 overflows uint8"},
		{`fn f(): int8 { return 300; }`, "integer constant 300 overflows int8"},
		{`fn f(): int8 { x: int64 := 1; return x; }`, "cannot use int64 value as int8"},
		{`s := []int8{1, 2, 300};`, "integer constant 300 overflows int8"},
		{`m := map[string]uint8{"a": 256};`, "integer constant 256 overflows uint8"},
		{`x := int8(300);`, "integer constant 300 overflows int8"},
		{`x := int8("a");`, "cannot convert string to int8"},
		{`x := int8(1, 2);`, "conversion to int8 expects 1 argument, got 2"},
		{`a: int32 := 1; b: int64 := 2; x := if (true) { a } else { b };`, "mismatched if branch types int32 and int64"},
		{`x: int8 := if (true) { 1 } else { 300 };`, "integer constant 300 overflows int8"},
		{`a: int32 := 1; b: int64 := 2; x := lbl: for { if (true) { break lbl (a); } break lbl (b); };`, "mismatched break value types int32 and int64"},
		{`s := []string{"a"}; i: int64 := 0; v := s[i];`, "cannot use int64 value as int"},
		{`m := map[uint8]string{}; v := m[300];`, "integer constant 300 overflows uint8"},
		{`struct P { pub age: int8; } p := P{age: 300};`, "integer constant 300 overflows int8"},
		{`struct P { age: int8 = 300; }`, "integer constant 300 overflows int8"},
		{`fn pair(): int32, int32 { return 1, 2; } mut a: int64 := 0; mut b: int32 := 0; (a, b) = pair();`, "cannot use int32 value as int64"},
		{`x := 9223372036854775808;`, "integer constant 9223372036854775808 overflows int"},
		{`x := true + 2;`, "cannot use untyped integer constant as bool"},
		{`x := 1 + true;`, "cannot use untyped integer constant as bool"},
		{`x := 1 + fn() {};`, "cannot use untyped integer constant as fn(): void"},
		{`struct Person { name: string; } p := Person{name: 1};`, "cannot use untyped integer constant as string"},
		{`struct Example { value: string = 1; }`, "cannot use untyped integer constant as string"},
		{`struct Person { name: string; fn new(name string): Person { return Person{name}; } } Person.new(1);`, "cannot use untyped integer constant as string"},
		{`x: int8 := -129;`, "integer constant -129 overflows int8"},
		{`x: int8 := 128;`, "integer constant 128 overflows int8"},
		{`x: uint8 := 256;`, "integer constant 256 overflows uint8"},
		{`x: int64 := -9223372036854775809;`, "integer constant -9223372036854775809 overflows int64"},
		{`x: int64 := 9223372036854775808;`, "integer constant 9223372036854775808 overflows int64"},
		{`x: uint64 := 18446744073709551616;`, "integer constant 18446744073709551616 overflows uint64"},
		{`x := uint8(-1);`, "integer constant -1 overflows uint8"},
		{`x := int32(true);`, "cannot convert bool to int32"},
		{`fn apply(f fn(int8): int8): int8 { return f(300); }`, "integer constant 300 overflows int8"},
		{`f: fn(uint8) := fn(n uint8) {}; f(-1);`, "integer constant -1 overflows uint8"},
		{`x := if (true) { 1 } else { true };`, "cannot use untyped integer constant as bool"},
		{`x := lbl: for { if (true) { break lbl ("a"); } break lbl (1); };`, "cannot use untyped integer constant as string"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			_, errs := parseAndInfer(t, test.input)
			if !containsError(errs, test.expected) {
				t.Fatalf("expected error containing %q, got %v", test.expected, errs)
			}
		})
	}
}

func TestInfer_LiteralsTakeTheirContextType(t *testing.T) {
	program, _ := parseAndInferProgram(t, `y: int32 := 10; w := y + 1;`)

	infix := program.Statements[1].(*ast.DeclareStatement).Value.(*ast.InfixExpression)
	literal := infix.Right.(*ast.IntegerLiteral)
	if literal.GetResolvedType() != types.Int32Type || infix.GetResolvedType() != types.Int32Type {
		t.Fatalf("expected the literal and sum to be int32, got %v and %v", literal.GetResolvedType(), infix.GetResolvedType())
	}
}

func TestInfer_EveryLiteralGetsAConcreteType(t *testing.T) {
	program, _ := parseAndInferProgram(t, `b := true; x := b.field + 1; y := b[1]; z := !5;`)

	for i, stmt := range program.Statements[1:] {
		var literal *ast.IntegerLiteral
		switch value := stmt.(*ast.DeclareStatement).Value.(type) {
		case *ast.InfixExpression:
			literal = value.Right.(*ast.IntegerLiteral)
		case *ast.IndexExpression:
			literal = value.Index.(*ast.IntegerLiteral)
		case *ast.PrefixExpression:
			literal = value.Right.(*ast.IntegerLiteral)
		}
		if literal.GetResolvedType() != types.IntType {
			t.Errorf("statement %d: literal has type %v, want int", i, literal.GetResolvedType())
		}
	}
}

func TestInfer_RangeVariables(t *testing.T) {
	program, _ := parseAndInferProgram(t, `for i, v := range []int32{1} {} for k, v := range map[string]uint8{} {} for kv := range []bool{} {}`)

	tests := []struct {
		stmt int
		want []string
	}{
		{0, []string{"int", "int32"}},
		{1, []string{"string", "uint8"}},
		{2, []string{"(int, bool)"}},
	}
	for _, test := range tests {
		loop := program.Statements[test.stmt].(*ast.ExpressionStatement).Expression.(*ast.ForLoop)
		for i, variable := range loop.Controller.(*ast.Range).Variables {
			if got := variable.GetResolvedType().Name(); got != test.want[i] {
				t.Errorf("loop %d variable %s has type %s, want %s", test.stmt, variable.Value, got, test.want[i])
			}
		}
	}
}

func TestInfer_NamedTypesGetTheirStruct(t *testing.T) {
	program, _ := parseAndInferProgram(t, `fn name(p Person): string { return p.name; } struct Person { pub name: string; }`)

	param := program.Statements[0].(*ast.FunctionStatement).Parameters[0].Type.(*types.Named)
	st, ok := param.Underlying.(*types.Struct)
	if !ok || st.Name() != "Person" {
		t.Fatalf("expected Person to have its struct as its underlying type, got %#v", param.Underlying)
	}
}

func TestInfer_ValidProgramsHaveNoErrors(t *testing.T) {
	tests := []string{
		`x := 0; y: int32 := 10; z := int32(x) + y;`,
		`for (i := 0; i < 10; i = i + 1) {}`,
		`s := []string{"a"}; for i := range s { x := s[i]; }`,
		`x := 1; if (x == 1) { 1; } else { "a"; }`,
		`fn fib(n uint64): uint64 { if (n < 2) { return n; } return fib(n - 1) + fib(n - 2); }`,
		`x := if (true) { 1; y := 2; } else { 3 };`,
		`x: int8 := -128;`,
		`x: int8 := 127;`,
		`x: uint8 := 0;`,
		`x: uint8 := 255;`,
		`x: int64 := -9223372036854775808;`,
		`x: int64 := 9223372036854775807;`,
		`x: uint64 := 18446744073709551615;`,
		`struct A { b: B; } struct B { pub n: int16 = 1; } a := A{b: B{}}; x := a.b.n + 1;`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			if _, errs := parseAndInfer(t, input); len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}
		})
	}
}

func TestInfer_StructFieldTypesFromDefaults(t *testing.T) {
	program, _ := parseAndInferProgram(t, `
		struct Address { city: string; }
		struct Profile {
			name = "Ada";
			score = 42;
			active = true;
			address = Address{city: "London"};
			tags = []string{"compiler"};
			explicit: int32 = 1;
		}
	`)

	profile := program.Statements[1].(*ast.StructStatement)
	expected := []string{"string", "int", "bool", "Address", "[]string", "int32"}
	for i, want := range expected {
		field := profile.Fields[i]
		if field.GetType() == nil || field.GetType().Name() != want {
			t.Errorf("field %s has type %v, want %s", field.Name.Value, field.GetType(), want)
		}
		if inferred := field.Name.Value != "explicit"; field.Inferred != inferred {
			t.Errorf("field %s: Inferred = %t, want %t", field.Name.Value, field.Inferred, inferred)
		}
	}
}

func TestInfer_NamedTypesInsideCollections(t *testing.T) {
	program, _ := parseAndInferProgram(t, `
		struct Team { captain: Person; members: []Person; history: [2]Person; }
		struct Person { name: string; }
	`)

	team := program.Statements[0].(*ast.StructStatement).Name.GetResolvedType().(*types.Struct)
	for field, want := range map[string]string{"captain": "Person", "members": "[]Person", "history": "[2]Person"} {
		typ := team.Fields[field].Type
		if typ.Name() != want {
			t.Errorf("field %s has type %s, want %s", field, typ.Name(), want)
		}
		var named *types.Named
		switch collection := typ.(type) {
		case *types.Named:
			named = collection
		case *types.Slice:
			named = collection.ElementType().(*types.Named)
		case *types.Array:
			named = collection.ElementType().(*types.Named)
		}
		if _, ok := named.Underlying.(*types.Struct); !ok {
			t.Errorf("field %s: Person has no underlying struct", field)
		}
	}
}

func TestInfer_RequiresResolvedProgram(t *testing.T) {
	p := parser.New(lexer.New(`x := 1;`))
	errs := New(p.ParseProgram()).Infer()
	if !containsError(errs, "inference requires a resolved program") {
		t.Fatalf("expected an error for an unresolved program, got %v", errs)
	}
}

func mustInfer(t *testing.T, input string) *resolver.Scope {
	t.Helper()
	scope, errs := parseAndInfer(t, input)
	if len(errs) > 0 {
		t.Fatalf("inference errors: %v", errs)
	}
	return scope
}

func parseAndInfer(t *testing.T, input string) (*resolver.Scope, []string) {
	t.Helper()
	program, errs := parseAndInferProgram(t, input)
	return program.Scope.(*resolver.Scope), errs
}

func parseAndInferProgram(t *testing.T, input string) (*ast.Program, []string) {
	t.Helper()
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	if _, errs := resolver.New(program).Resolve(); len(errs) > 0 {
		t.Fatalf("resolver errors: %v", errs)
	}
	return program, New(program).Infer()
}

func typeOfSymbol(t *testing.T, scope *resolver.Scope, name string) string {
	t.Helper()
	sym, ok := scope.GetLocal(name)
	if !ok {
		t.Fatalf("symbol %q not found", name)
	}
	typ := sym.Ident.GetResolvedType()
	if typ == nil {
		return "<nil>"
	}
	return typ.Name()
}

func containsError(errs []string, expected string) bool {
	for _, err := range errs {
		if strings.Contains(err, expected) {
			return true
		}
	}
	return false
}
