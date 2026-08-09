package parser

import (
	"testing"

	"github.com/esweby/primordial_lang/ast"
)

func TestParseStructLiteralVariants(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		structName string
		fieldCount int
	}{
		{name: "empty", input: `Empty{};`, structName: "Empty", fieldCount: 0},
		{name: "one shorthand", input: `Person{age};`, structName: "Person", fieldCount: 1},
		{name: "shorthand trailing comma", input: `Person{age,};`, structName: "Person", fieldCount: 1},
		{name: "explicit", input: `Person{age: 24, name: "Brian"};`, structName: "Person", fieldCount: 2},
		{name: "mixed", input: `Person{age, name: fullName,};`, structName: "Person", fieldCount: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			literal := requireStructLiteral(t, parseSingleStructExpression(t, test.input))
			if literal.Name.Value != test.structName {
				t.Fatalf("expected literal name %s, got %s", test.structName, literal.Name.Value)
			}
			if len(literal.Fields) != test.fieldCount {
				t.Fatalf("expected %d fields, got %d", test.fieldCount, len(literal.Fields))
			}
		})
	}
}

func TestParseExplicitStructLiteralValues(t *testing.T) {
	literal := requireStructLiteral(t, parseSingleStructExpression(t,
		`Person{age: currentAge + 1, name: format(firstName, lastName)};`,
	))

	age := literal.Fields[0]
	if age.Shorthand {
		t.Fatal("expected age to be an explicit field")
	}
	if !testInfixExpression(t, age.Value, "currentAge", "+", 1) {
		t.FailNow()
	}

	name := literal.Fields[1]
	call, ok := name.Value.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected name value to be a CallExpression, got %T", name.Value)
	}
	if call.Function.String() != "format" {
		t.Fatalf("expected format call, got %s", call.Function.String())
	}
	if len(call.Arguments) != 2 {
		t.Fatalf("expected 2 call arguments, got %d", len(call.Arguments))
	}
}

func TestParseNestedStructLiteral(t *testing.T) {
	literal := requireStructLiteral(t, parseSingleStructExpression(t,
		`Person{name: "Brian", address: Address{city: "London"}};`,
	))

	address := literal.Fields[1]
	nested := requireStructLiteral(t, address.Value)
	if nested.Name.Value != "Address" {
		t.Fatalf("expected nested Address literal, got %s", nested.Name.Value)
	}
	if len(nested.Fields) != 1 || nested.Fields[0].Name.Value != "city" {
		t.Fatalf("unexpected nested fields: %+v", nested.Fields)
	}
}

func parseSingleStructExpression(t *testing.T, input string) ast.Expression {
	t.Helper()
	program, parser := parseStructTestProgram(input)
	requireNoParserErrors(t, parser)
	requireStatementCount(t, program.Statements, 1)

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
	}
	return statement.Expression
}

func requireStructLiteral(t *testing.T, expression ast.Expression) *ast.StructLiteral {
	t.Helper()
	literal, ok := expression.(*ast.StructLiteral)
	if !ok {
		t.Fatalf("expected StructLiteral, got %T", expression)
	}
	return literal
}
