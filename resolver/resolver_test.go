package resolver

import (
	"strings"
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/parser"
	"github.com/esweby/primordial_lang/types"
)

func TestResolver_DeclareStatement(t *testing.T) {
	program, scope := mustResolve(t, `x := 0;`)

	sym, ok := scope.GetLocal("x")
	if !ok {
		t.Fatalf("symbol %q not found in scope", "x")
	}
	if sym.Kind != VariableSymbol {
		t.Errorf("sym.Kind = %s, want %s", sym.Kind, VariableSymbol)
	}
	if sym.Declaration() != program.Statements[0] {
		t.Errorf("symbol declaration is %T, want the declare statement", sym.Declaration())
	}
	if program.Scope != ast.Scope(scope) {
		t.Errorf("program.Scope was not set to the returned scope")
	}
}

func TestResolver_ValidPrograms(t *testing.T) {
	tests := []string{
		`x := 1; y := x + 1;`,
		`mut x := 1; x = 2;`,
		`mut a := 1; mut b := 2; (a, b) = (b, a);`,
		`fn pair(): int32, bool { return 1, true; } (a, _) := pair();`,
		`fn fact(n int64): int64 { if (n == 0) { return 1; } return n * fact(n - 1); }`,
		`f := fn(n int64): int64 { return f(n); };`,
		`x := 1; if (x == 1) { x := 2; } else { x := 3; }`,
		`x := 1; y := if (x == 1) { 2 } else if (x == 2) { 3 } else { 4 };`,
		`for {}`,
		`for (1 < 2) {}`,
		`for (i := 1; i < 10; i = i + 1) { i; }`,
		`for a, b := range map[int32]string{} { a; b; }`,
		`for kv := range []int32{1, 2} { kv; }`,
		`for { break; }`,
		`for { continue; }`,
		`lbl: for { break lbl; }`,
		`outer: for { for { break outer; } }`,
		`x := lbl: for { break lbl (1); }`,
		`s := []int32{1, 2}; n := len(s); i := int32(n);`,
		`len := 1; y := len;`,
		`m := map[string]int32{"a": 1}; v := m["a"];`,
		`struct Person { name: string; } p := Person{name: "Ada"}; n := p.name;`,
		`struct Person { name: string; } name := "Ada"; p := Person{name};`,
		`p := Person{name: "Ada"}; struct Person { name: string; }`,
		`struct Team { captain: Person; members: []Person; history: [2]Person; } struct Person { name: string; }`,
		`fn greet(p Person): string { return p.name; } struct Person { pub name: string; }`,
		`struct Person { name: string; impl { fn getName(): string { return self.name; } } }`,
		`struct Person { name: string; fn new(name string): Person { return Person{name: name}; } } p := Person.new("Ada");`,
		`struct Counter { count: int32 = 0; }`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, errs := parseAndResolve(t, input)
			if len(errs) > 0 {
				t.Fatalf("unexpected resolver errors: %v", errs)
			}
		})
	}
}

func TestResolver_Errors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`y := x;`, "undefined identifier: x"},
		{`x = 1;`, "undefined identifier: x"},
		{`fn pair(): int32, int32 { return 1, 2; } (a, b) = pair();`, "undefined identifier: a"},
		{`x := x + 1;`, "undefined identifier: x"},
		{`x := 1; x := 2;`, "variable 'x' already declared in this scope"},
		{`fn f() {} fn f() {}`, "function 'f' already declared in this scope"},
		{`fn f(a int32, a int32) {}`, "parameter 'a' already declared in this scope"},
		{`fn f(a int32) { a := 1; }`, "parameter 'a' already declared in this scope"},
		{`struct A {} struct A {}`, "struct 'A' already declared in this scope"},
		{`fn pair(): int32, int32 { return 1, 2; } (a, a) := pair();`, "variable 'a' appears more than once in tuple declaration"},
		{`if (true) { y := 1; } z := y;`, "undefined identifier: y"},
		{`for (i := 0; i < 1; i = i + 1) {} j := i;`, "undefined identifier: i"},
		{`fn f(a int32) {} b := a;`, "undefined identifier: a"},
		{`fn f() { g(); } fn g() {}`, "undefined identifier: g"},
		{`break;`, "break outside of loop"},
		{`continue;`, "continue outside of loop"},
		{`if (true) { break; }`, "break outside of loop"},
		{`for { f := fn() { break; }; }`, "break outside of loop"},
		{`a: for { break b; }`, "break label 'b' does not match any enclosing loop"},
		{`a: for {} for { break a; }`, "break label 'a' does not match any enclosing loop"},
		{`x: Missing := 1;`, "unknown type: Missing"},
		{`fn f(p Missing) {}`, "unknown type: Missing"},
		{`fn f(): Missing {}`, "unknown type: Missing"},
		{`struct A { b: []Missing; }`, "unknown type: Missing"},
		{`m := map[string]Missing{};`, "unknown type: Missing"},
		{`Person := 1; x: Person := 1;`, "Person is not a type"},
		{`p := Missing{};`, "undefined identifier: Missing"},
		{`Person := 1; p := Person{};`, "Person is not a struct type"},
		{`struct A { b: int32 = c; }`, "undefined identifier: c"},
		{`struct A { impl { fn f(): int32 { return missing; } } }`, "undefined identifier: missing"},
		{`fn f() { self; }`, "undefined identifier: self"},
		{`struct A { fn new(): A { return self; } }`, "undefined identifier: self"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			_, errs := parseAndResolve(t, test.input)
			if !containsError(errs, test.expected) {
				t.Fatalf("expected error containing %q, got %v", test.expected, errs)
			}
		})
	}
}

func TestResolver_ErrorsIncludePosition(t *testing.T) {
	_, errs := parseAndResolve(t, "x := 1;\ny := z;")
	if len(errs) != 1 || !strings.HasPrefix(errs[0], "2:6: ") {
		t.Fatalf("expected a single error at 2:6, got %v", errs)
	}
}

func TestResolver_IdentifierLinksToDeclaration(t *testing.T) {
	program, _ := mustResolve(t, `x := 1; y := x;`)

	decl := program.Statements[0].(*ast.DeclareStatement)
	use := program.Statements[1].(*ast.DeclareStatement).Value.(*ast.Identifier)

	if use.Symbol == nil {
		t.Fatal("identifier was not resolved")
	}
	if use.Symbol.Declaration() != decl {
		t.Fatalf("identifier resolved to %T, want the declaration of x", use.Symbol.Declaration())
	}
	if decl.Name.Symbol != use.Symbol {
		t.Fatal("declaring identifier and use should share a symbol")
	}
}

func TestResolver_ShadowingResolvesToInnermost(t *testing.T) {
	program, scope := mustResolve(t, `x := 1; if (true) { x := 2; y := x; } z := x;`)

	outer := program.Statements[0].(*ast.DeclareStatement)
	ifExpr := program.Statements[1].(*ast.ExpressionStatement).Expression.(*ast.IfExpression)
	inner := ifExpr.Body.Statements[0].(*ast.DeclareStatement)
	innerUse := ifExpr.Body.Statements[1].(*ast.DeclareStatement).Value.(*ast.Identifier)
	outerUse := program.Statements[2].(*ast.DeclareStatement).Value.(*ast.Identifier)

	if innerUse.Symbol.Declaration() != inner {
		t.Error("inner use of x should resolve to the inner declaration")
	}
	if outerUse.Symbol.Declaration() != outer {
		t.Error("outer use of x should resolve to the outer declaration")
	}

	if ifExpr.Body.Scope == nil || ifExpr.Body.Scope.Parent() != ast.Scope(scope) {
		t.Error("if body scope should be a child of the program scope")
	}
	all, _ := scope.FindAllSymbols("x")
	if len(all) != 2 {
		t.Errorf("FindAllSymbols(x) found %d symbols, want 2", len(all))
	}
}

func TestResolver_FunctionScopes(t *testing.T) {
	program, scope := mustResolve(t, `fn add(a int32, b int32): int32 { return a + b; }`)

	fn := program.Statements[0].(*ast.FunctionStatement)
	sym, ok := scope.GetLocal("add")
	if !ok || sym.Kind != FunctionSymbol || sym.Declaration() != fn {
		t.Fatalf("expected add to be a function symbol, got %#v", sym)
	}
	if fn.Scope == nil || fn.Body.Scope != fn.Scope {
		t.Fatal("function and its body should share one scope")
	}
	if sym.Body != fn.Scope {
		t.Fatal("function symbol should own the function scope")
	}

	ret := fn.Body.Statements[0].(*ast.ReturnStatement)
	a := ret.ReturnValues[0].(*ast.InfixExpression).Left.(*ast.Identifier)
	if a.Symbol.Declaration() != fn.Parameters[0] {
		t.Fatalf("a resolved to %T, want the first parameter", a.Symbol.Declaration())
	}
}

func TestResolver_MethodSelf(t *testing.T) {
	program, _ := mustResolve(t, `struct Person { name: string; impl { fn getName(): string { return self.name; } } }`)

	st := program.Statements[0].(*ast.StructStatement)
	method := st.Impl.Methods[0]
	ret := method.Body.Statements[0].(*ast.ReturnStatement)
	self := ret.ReturnValues[0].(*ast.MemberExpression).Receiver.(*ast.Identifier)

	sym, ok := self.Symbol.(*Symbol)
	if !ok || sym.Kind != SelfSymbol || sym.Declaration() != st {
		t.Fatalf("self should resolve to the struct's receiver, got %#v", self.Symbol)
	}
}

func TestResolver_RangeVariables(t *testing.T) {
	program, _ := mustResolve(t, `for k, v := range map[string]int32{} { k; v; }`)

	loop := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.ForLoop)
	if loop.Scope == nil || loop.Body.Scope != loop.Scope {
		t.Fatal("loop and its body should share one scope")
	}

	use := loop.Body.Statements[1].(*ast.ExpressionStatement).Expression.(*ast.Identifier)
	sym := use.Symbol.(*Symbol)
	if sym.Kind != RangeSymbol || sym.Declaration() != loop.Controller {
		t.Fatalf("v should resolve to the range controller, got %#v", sym)
	}
}

func TestResolver_Builtins(t *testing.T) {
	program, scope := mustResolve(t, `n := len([]int32{});`)

	call := program.Statements[0].(*ast.DeclareStatement).Value.(*ast.CallExpression)
	sym := call.Function.(*ast.Identifier).Symbol.(*Symbol)
	if sym.Kind != BuiltinSymbol || sym.Declaration() != nil {
		t.Fatalf("len should resolve to a builtin, got %#v", sym)
	}
	if sym.Scope != scope.ParentScope() || scope.Parent() == nil || scope.ParentScope().Parent() != nil {
		t.Fatal("builtins should live in the universe scope above the program scope")
	}
}

func TestResolver_NamedTypesLinkToStruct(t *testing.T) {
	program, _ := mustResolve(t, `
		struct Team { captain: Person; members: []Person; }
		fn lead(team Team): Person { return team.captain; }
		struct Person { name: string; }
	`)

	team := program.Statements[0].(*ast.StructStatement)
	lead := program.Statements[1].(*ast.FunctionStatement)
	person := program.Statements[2].(*ast.StructStatement)

	named := []types.Type{
		team.Fields[0].Type,
		team.Fields[1].Type.(*types.Slice).ElementType(),
		lead.ReturnTypes[0].Type,
	}
	for _, typ := range named {
		sym, ok := typ.(*types.Named).Symbol.(*Symbol)
		if !ok || sym.Declaration() != person {
			t.Errorf("%s should link to the Person declaration, got %#v", typ.Name(), typ.(*types.Named).Symbol)
		}
	}

	param := lead.Parameters[0].Type.(*types.Named).Symbol.(*Symbol)
	if param.Declaration() != team {
		t.Errorf("parameter type should link to the Team declaration, got %#v", param)
	}
}

func mustResolve(t *testing.T, input string) (*ast.Program, *Scope) {
	t.Helper()
	program := parse(t, input)
	scope, errs := New(program).Resolve()
	if len(errs) > 0 {
		t.Fatalf("resolver errors: %v", errs)
	}
	return program, scope
}

func parseAndResolve(t *testing.T, input string) (*Scope, []string) {
	t.Helper()
	return New(parse(t, input)).Resolve()
}

func parse(t *testing.T, input string) *ast.Program {
	t.Helper()
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	return program
}

func containsError(errs []string, expected string) bool {
	for _, err := range errs {
		if strings.Contains(err, expected) {
			return true
		}
	}
	return false
}
