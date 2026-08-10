package parser

import (
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/types"
)

func TestParseEmptyStructDeclaration(t *testing.T) {
	program, parser := parseStructTestProgram(`struct Empty {}`)
	requireNoParserErrors(t, parser)
	requireStatementCount(t, program.Statements, 1)

	declaration := requireStructStatement(t, program.Statements[0])
	if declaration.Name.Value != "Empty" {
		t.Fatalf("expected struct name Empty, got %s", declaration.Name.Value)
	}
	if len(declaration.Fields) != 0 {
		t.Fatalf("expected no fields, got %d", len(declaration.Fields))
	}
	if len(declaration.TypeFunctions) != 0 {
		t.Fatalf("expected no type functions, got %d", len(declaration.TypeFunctions))
	}
	if declaration.Impl != nil {
		t.Fatalf("expected no impl block, got %T", declaration.Impl)
	}
}

func TestParseStructFieldDeclarations(t *testing.T) {
	input := `struct Profile {
		pub displayName: string;
		score: int64 = 100;
		visits: uint32 = 0;
		owner: Person;
		friends: []Person;
		history: [3]Person;
	}`
	program, parser := parseStructTestProgram(input)
	requireNoParserErrors(t, parser)

	declaration := requireStructStatement(t, program.Statements[0])
	if len(declaration.Fields) != 6 {
		t.Fatalf("expected 6 fields, got %d", len(declaration.Fields))
	}

	tests := []struct {
		name       string
		typeName   string
		public     bool
		hasDefault bool
	}{
		{name: "displayName", typeName: "string", public: true},
		{name: "score", typeName: "int64", hasDefault: true},
		{name: "visits", typeName: "uint32", hasDefault: true},
		{name: "owner", typeName: "Person"},
		{name: "friends", typeName: "[]Person"},
		{name: "history", typeName: "[3]Person"},
	}

	for i, expected := range tests {
		field := declaration.Fields[i]
		if field.Name.Value != expected.name {
			t.Errorf("field %d: expected name %s, got %s", i, expected.name, field.Name.Value)
		}
		if field.Type.Name() != expected.typeName {
			t.Errorf("field %d: expected type %s, got %s", i, expected.typeName, field.Type.Name())
		}
		if field.Public != expected.public {
			t.Errorf("field %d: expected Public=%t, got %t", i, expected.public, field.Public)
		}
		if (field.Value != nil) != expected.hasDefault {
			t.Errorf("field %d: expected hasDefault=%t, got %t", i, expected.hasDefault, field.Value != nil)
		}
	}

	if !testIntegerLiteral(t, declaration.Fields[1].Value, 100) {
		t.FailNow()
	}

	owner, ok := declaration.Fields[3].Type.(*types.Named)
	if !ok {
		t.Fatalf("expected owner to use a named type, got %T", declaration.Fields[3].Type)
	}
	if owner.Underlying != types.InvalidType {
		t.Fatalf("expected unresolved named type, got underlying type %s", owner.Underlying.Name())
	}
}

func TestParseStructDeclarationWithTypeFunctionOnly(t *testing.T) {
	input := `struct Coordinate {
		x: int32;
		y: int32;

		fn origin(): Coordinate {
			return Coordinate{x: 0, y: 0};
		}
	}`
	program, parser := parseStructTestProgram(input)
	requireNoParserErrors(t, parser)

	declaration := requireStructStatement(t, program.Statements[0])
	if declaration.Impl != nil {
		t.Fatal("expected struct without an impl block")
	}
	if len(declaration.TypeFunctions) != 1 {
		t.Fatalf("expected 1 type function, got %d", len(declaration.TypeFunctions))
	}

	function := declaration.TypeFunctions[0]
	if function.Name.Value != "origin" {
		t.Fatalf("expected origin function, got %s", function.Name.Value)
	}
	if len(function.ReturnTypes) != 1 || function.ReturnTypes[0].Type.Name() != "Coordinate" {
		t.Fatalf("expected Coordinate return type, got %+v", function.ReturnTypes)
	}
	if len(function.Body.Statements) != 1 {
		t.Fatalf("expected one function statement, got %d", len(function.Body.Statements))
	}
}

func TestRejectMutableStructFieldModifier(t *testing.T) {
	_, parser := parseStructTestProgram(`struct Person { mut name: string; }`)
	diagnostics := parser.Diagnostics()
	if len(diagnostics) == 0 {
		t.Fatal("expected mut struct field modifier to be rejected")
	}
	if diagnostics[0].Code != "P1304" || diagnostics[0].Message != "struct fields only support the pub modifier" {
		t.Fatalf("unexpected parser diagnostic: %+v", diagnostics[0])
	}
}

func parseStructTestProgram(input string) (*ast.Program, *Parser) {
	parser := New(lexer.New(input))
	return parser.ParseProgram(), parser
}

func requireStructStatement(t *testing.T, statement ast.Statement) *ast.StructStatement {
	t.Helper()
	declaration, ok := statement.(*ast.StructStatement)
	if !ok {
		t.Fatalf("expected StructStatement, got %T", statement)
	}
	return declaration
}
