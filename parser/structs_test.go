package parser

import (
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/types"
)

func TestStructDeclareStatement(t *testing.T) {
	input := `struct Person {
	age: int32;
	name: string;

	pub fn new(age int32, firstName string, lastName string): Person {
		return Person{
			age,
			name: firstName + " " + lastName,
		}
	}

	impl {
		fn getName(): string { return self.name; }

		fn incrementAge() { self.age = self.age + 1; }
	}
}`;

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	requireNoParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	requireStatementCount(t, program.Statements, 1)

	stmt := program.Statements[0]

	if stmt.TokenLiteral() != "struct" {
		t.Errorf("stmt.TokenLiteral not 'struct'. Got=%q", stmt.TokenLiteral())
	}

	structStmt, ok := stmt.(*ast.StructStatement)
	if !ok {
		t.Fatalf("stmt not *ast.StructStatement. Got=%T", stmt)
	}

	if structStmt.Name.Value != "Person" {
		t.Errorf("structStmt.Name.Value not 'Person'. Got=%s", structStmt.Name.Value)
	}

	fieldTests := []struct{
		name string
		t types.Type
	}{ 
		{ "age", types.Int32Type },
		{ "name", types.StringType },
	}

	fields := structStmt.Fields
	if len(fields) != len(fieldTests) {
		t.Fatalf("stmt.Fields length not '2'. Got=%d", len(fields))
	}

	for i, tt := range fieldTests {
		f := fields[i]
		if tt.name != f.Name.Value {
			t.Errorf(
				"field test %d failed: expected name '%s', got=%s",
				i,
				tt.name,
				f.Name.Value,
			)
		}

		if !types.IsTypesEqual(tt.t, f.Type) {
			t.Errorf(
				"field test %d failed: expected type '%s', got=%s",
				i,
				tt.t.Name(),
				f.Type.Name(),
			)
		}
	}

	
	typeFnTests := []struct{
		name string
		numParams int
		numReturn int
		numStatements int
	}{
		{ "new", 3, 1, 1 },
	}
		
	typeFns := structStmt.TypeFunctions
	if len(typeFns) != len(typeFnTests) {
		t.Fatalf("stmt.typeFns length not '1'. Got=%d", len(typeFns))
	}

	for i, tt := range typeFnTests {
		fn := typeFns[i]
		if fn.Name.Value != tt.name {
			t.Errorf("typeFnTest %d failed: expected %s name. got=%s", i, 
				tt.name, fn.Name.Value)
		}

		if len(fn.Parameters) != tt.numParams {
			t.Errorf("typeFnTest %d failed: expected %d parameters. got=%d", i, tt.numParams, len(fn.Parameters))
		}

		if len(fn.ReturnTypes) != tt.numReturn {
			t.Errorf("typeFnTest %d failed: expected %d returnTypes. got=%d", i, tt.numReturn, len(fn.ReturnTypes))
		}

		if len(fn.Body.Statements) != tt.numStatements {
			t.Errorf("typeFnTest %d failed: expected %d statements. got=%d", i, tt.numStatements, len(fn.Body.Statements))
		}
	}

	if structStmt.Impl == nil {
		t.Fatalf("structStmt.Impl is nil, expected to be ast.StructImplBlock")
	}

	implMethodTests := []struct{
		name string
		numParams int
		numReturn int
		numStatements int
	}{
		{ "getName", 0, 1, 1 },
		{ "incrementAge", 0, 0, 1 },
	}
	
	implMethods := structStmt.Impl.Methods
	if len(implMethods) != len(implMethodTests) {
		t.Fatalf("stmt.typeFns length not '2'. Got=%d", len(implMethods))
	}

	for i, tt := range implMethodTests {
		fn := implMethods[i]
		if fn.Name.Value != tt.name {
			t.Errorf("implMethodTests %d failed: expected %s name. got=%s", i, 
				tt.name, fn.Name.Value)
		}

		if len(fn.Parameters) != tt.numParams {
			t.Errorf("implMethodTests %d failed: expected %d parameters. got=%d", i, tt.numParams, len(fn.Parameters))
		}

		if len(fn.ReturnTypes) != tt.numReturn {
			t.Errorf("implMethodTests %d failed: expected %d returnTypes. got=%d", i, tt.numReturn, len(fn.ReturnTypes))
		}

		if len(fn.Body.Statements) != tt.numStatements {
			t.Errorf("implMethodTests %d failed: expected %d statements. got=%d", i, tt.numStatements, len(fn.Body.Statements))
		}
	}
	
}