package parser

import (
	"testing"

	"github.com/esweby/primordial_lang/ast"
)

func TestParseStructFieldAccess(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "field", input: `person.name;`, expected: "person.name"},
		{name: "nested field", input: `person.address.city;`, expected: "person.address.city"},
		{name: "field after method call", input: `person.getAddress().city;`, expected: "person.getAddress().city"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expression := parseSingleStructExpression(t, test.input)
			member, ok := expression.(*ast.MemberExpression)
			if !ok {
				t.Fatalf("expected MemberExpression, got %T", expression)
			}
			if member.String() != test.expected {
				t.Fatalf("expected %s, got %s", test.expected, member.String())
			}
		})
	}
}

func TestParseStructLiteralFieldAccess(t *testing.T) {
	expression := parseSingleStructExpression(t, `Person{name: "Brian"}.name;`)
	member, ok := expression.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("expected MemberExpression, got %T", expression)
	}

	receiver := requireStructLiteral(t, member.Receiver)
	if receiver.Name.Value != "Person" {
		t.Fatalf("expected Person receiver, got %s", receiver.Name.Value)
	}
	if member.Name.Value != "name" {
		t.Fatalf("expected name member, got %s", member.Name.Value)
	}
}

func TestParseStructFieldAccessInReturn(t *testing.T) {
	program, parser := parseStructTestProgram(`return person.name;`)
	requireNoParserErrors(t, parser)
	requireStatementCount(t, program.Statements, 1)

	statement, ok := program.Statements[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("expected ReturnStatement, got %T", program.Statements[0])
	}
	if len(statement.ReturnValues) != 1 {
		t.Fatalf("expected one return value, got %d", len(statement.ReturnValues))
	}
	if statement.ReturnValues[0].String() != "person.name" {
		t.Fatalf("expected person.name return value, got %s", statement.ReturnValues[0].String())
	}
}

func TestParseNestedStructFieldAssignment(t *testing.T) {
	program, parser := parseStructTestProgram(`person.address.city = "London";`)
	requireNoParserErrors(t, parser)
	requireStatementCount(t, program.Statements, 1)

	assignment, ok := program.Statements[0].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", program.Statements[0])
	}
	if assignment.Target.String() != "person.address.city" {
		t.Fatalf("expected person.address.city target, got %s", assignment.Target.String())
	}
	if assignment.Value.String() != "London" {
		t.Fatalf("expected London value, got %s", assignment.Value.String())
	}
}
