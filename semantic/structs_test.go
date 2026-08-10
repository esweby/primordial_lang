package semantic

import (
	"strings"
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/parser"
	"github.com/esweby/primordial_lang/types"
)

func analyzeProgram(t *testing.T, input string) (*SemanticAnalyzer, []error) {
	t.Helper()
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	analyzer := NewSemanticAnalyzer(program, NewSymbolTable())
	return analyzer, analyzer.Analyze()
}

func TestAnalyzeBasicStructs(t *testing.T) {
	analyzer, errors := analyzeProgram(t, `
		struct Address { pub city: string; }
		struct Person { name: string; pub address: Address; age: int64 = 42; }
		person := Person{name: "Ada", address: Address{city: "London"}};
		city := person.address.city;
	`)
	if len(errors) != 0 {
		t.Fatalf("expected valid structs, got %v", errors)
	}

	city, ok := analyzer.Symbols().Get("city")
	if !ok || city.Type().Name() != "string" {
		t.Fatalf("expected city to be a string, got %#v", city)
	}
}

func TestRejectInvalidStructLiterals(t *testing.T) {
	tests := []struct {
		name     string
		literal  string
		expected string
	}{
		{name: "missing", literal: `Person{}`, expected: "missing required field Person.name"},
		{name: "unknown", literal: `Person{name: "Ada", extra: 1}`, expected: "type Person has no field extra"},
		{name: "wrong type", literal: `Person{name: 1}`, expected: "field Person.name: cannot use untyped integer as string"},
		{name: "duplicate", literal: `Person{name: "Ada", name: "Grace"}`, expected: "field name supplied more than once"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, errors := analyzeProgram(t, `struct Person { name: string; } `+test.literal)
			if len(errors) == 0 {
				t.Fatal("expected a semantic error")
			}
			found := false
			for _, err := range errors {
				if strings.Contains(err.Error(), test.expected) {
					found = true
				}
			}
			if !found {
				t.Fatalf("expected error containing %q, got %v", test.expected, errors)
			}
		})
	}
}

func TestInferStructFieldTypesFromDefaults(t *testing.T) {
	p := parser.New(lexer.New(`
		struct Address { city: string; }
		struct Profile {
			name = "Ada";
			score = 42;
			active = true;
			address = Address{city: "London"};
			tags = []string{"compiler"};
			explicit: int32 = 1;
		}
	`))
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	analyzer := NewSemanticAnalyzer(program, NewSymbolTable())
	if errors := analyzer.Analyze(); len(errors) != 0 {
		t.Fatalf("expected field inference to succeed, got %v", errors)
	}

	declaration, ok := program.Statements[1].(*ast.StructStatement)
	if !ok {
		t.Fatalf("expected StructStatement, got %T", program.Statements[1])
	}
	expectedTypes := []string{"string", "int64", "bool", "Address", "[]string", "int32"}
	for i, expected := range expectedTypes {
		field := declaration.Fields[i]
		if field.GetType() == nil || field.GetType().Name() != expected {
			t.Errorf("field %s: expected inferred type %s, got %v", field.Name.Value, expected, field.GetType())
		}
		shouldInfer := field.Name.Value != "explicit"
		if field.Inferred != shouldInfer {
			t.Errorf("field %s: expected Inferred=%t, got %t", field.Name.Value, shouldInfer, field.Inferred)
		}
	}
}

func TestResolveNamedStructTypesInsideCollections(t *testing.T) {
	analyzer, errors := analyzeProgram(t, `
		struct Team {
			captain: Person;
			members: []Person;
			history: [2]Person;
		}
		struct Person { name: string; }
	`)
	if len(errors) != 0 {
		t.Fatalf("expected forward and collection type references to resolve, got %v", errors)
	}

	symbol, ok := analyzer.Symbols().Get("Team")
	if !ok {
		t.Fatal("expected Team type symbol")
	}
	team, ok := symbol.Type().(*types.Struct)
	if !ok {
		t.Fatalf("expected struct type, got %T", symbol.Type())
	}
	for field, expected := range map[string]string{
		"captain": "Person",
		"members": "[]Person",
		"history": "[2]Person",
	} {
		if actual := team.Fields[field].Type.Name(); actual != expected {
			t.Errorf("field %s: expected %s, got %s", field, expected, actual)
		}
	}
}

func TestRejectInvalidStructFieldDeclarations(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected string
	}{
		{name: "missing type and default", field: `value;`, expected: "requires a type or default value"},
		{name: "unknown type", field: `value: Missing;`, expected: "unknown type Missing"},
		{name: "default mismatch", field: `value: string = 1;`, expected: "cannot use untyped integer as string"},
		{name: "duplicate field", field: `value: int64; value: int64;`, expected: "field 'value' already declared"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, errors := analyzeProgram(t, `struct Example { `+test.field+` }`)
			if !containsSemanticError(errors, test.expected) {
				t.Fatalf("expected error containing %q, got %v", test.expected, errors)
			}
		})
	}
}

func TestStructTypeCannotBeUsedAsValue(t *testing.T) {
	_, errors := analyzeProgram(t, `struct Person {} value := Person;`)
	if !containsSemanticError(errors, "type Person cannot be used as a value") {
		t.Fatalf("expected type-as-value error, got %v", errors)
	}
}

func TestAnalyzeStructLiteralShorthand(t *testing.T) {
	_, errors := analyzeProgram(t, `
		struct Person { name: string; }
		name := "Ada";
		person := Person{name};
	`)
	if len(errors) != 0 {
		t.Fatalf("expected shorthand field to use the variable type, got %v", errors)
	}
}

func TestAnalyzeStructTypeFunctionsAndInstanceMethods(t *testing.T) {
	analyzer, errors := analyzeProgram(t, `
		struct Person {
			name: string;
			pub age: int32;

			fn new(name string, age int32): Person {
				return Person{name, age};
			}

			impl {
				fn getName(): string { return self.name; }
				fn hasName(other Person): bool { return self.name == other.name; }
				fn birthday(): int32 {
					self.age = self.age + 1;
					return self.age;
				}
			}
		}

		person := Person.new("Ada", 41);
		other := Person.new("Ada", 30);
		name := person.getName();
		matches := person.hasName(other);
		age := person.birthday();
	`)
	if len(errors) != 0 {
		t.Fatalf("expected struct functions and methods to type check, got %v", errors)
	}

	name, ok := analyzer.Symbols().Get("name")
	if !ok || !types.IsTypesEqual(name.Type(), types.StringType) {
		t.Fatalf("expected method result name to be string, got %#v", name)
	}
	age, ok := analyzer.Symbols().Get("age")
	if !ok || !types.IsTypesEqual(age.Type(), types.Int32Type) {
		t.Fatalf("expected method result age to be int32, got %#v", age)
	}
	matches, ok := analyzer.Symbols().Get("matches")
	if !ok || !types.IsTypesEqual(matches.Type(), types.BoolType) {
		t.Fatalf("expected private field comparison to return bool, got %#v", matches)
	}
}

func TestRejectStructMemberMisuse(t *testing.T) {
	tests := []struct {
		name     string
		program  string
		expected string
	}{
		{
			name:     "external field assignment",
			program:  `struct Person { pub name: string; } person := Person{name: "Ada"}; person.name = "Grace";`,
			expected: "cannot assign to field outside its struct: name",
		},
		{
			name:     "wrong type function argument",
			program:  `struct Person { name: string; fn new(name string): Person { return Person{name}; } } Person.new(1);`,
			expected: "argument 0 to new",
		},
		{
			name:     "instance method on type",
			program:  `struct Person { name: string; impl { fn getName(): string { return self.name; } } } Person.getName();`,
			expected: "type Person has no type function getName",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, errors := analyzeProgram(t, test.program)
			if !containsSemanticError(errors, test.expected) {
				t.Fatalf("expected error containing %q, got %v", test.expected, errors)
			}
		})
	}
}

func TestStructMethodMayAssignItsField(t *testing.T) {
	_, errors := analyzeProgram(t, `
		struct Person {
			pub name: string;
			impl { fn setName(name string) { self.name = name; } }
		}
		person := Person{name: "Ada"};
		person.setName("Grace");
	`)
	if len(errors) != 0 {
		t.Fatalf("expected internal field assignment to type check, got %v", errors)
	}
}

func TestAnalyzeNestedStructMethodMutation(t *testing.T) {
	analyzer, errors := analyzeProgram(t, `
		struct Address {
			pub city: string;
			impl { fn setCity(city string) { self.city = city; } }
		}
		struct Person { pub address: Address; }
		person := Person{address: Address{city: "London"}};
		person.address.setCity("Paris");
		city := person.address.city;
	`)
	if len(errors) != 0 {
		t.Fatalf("expected nested struct method mutation to type check, got %v", errors)
	}
	city, ok := analyzer.Symbols().Get("city")
	if !ok || !types.IsTypesEqual(city.Type(), types.StringType) {
		t.Fatalf("expected nested city to be string, got %#v", city)
	}
}

func TestRejectPrivateStructFieldAccess(t *testing.T) {
	_, errors := analyzeProgram(t, `
		struct Person { name: string; }
		person := Person{name: "Ada"};
		person.name;
	`)
	if !containsSemanticError(errors, "field Person.name is private") {
		t.Fatalf("expected private field error, got %v", errors)
	}
}

func containsSemanticError(errors []error, expected string) bool {
	for _, err := range errors {
		if strings.Contains(err.Error(), expected) {
			return true
		}
	}
	return false
}
