package parser

import (
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/types"
)

func TestMapLiteral(t *testing.T) {
	tests := []struct{
		input string
		expectedKeyType types.Type
		expectedValueType types.Type
		expectedKey string
		expectedValue string
	}{
		{`map[int32]string{ 10: "ten", }`, types.Int32Type, types.StringType, "10", "ten"},
		{`map[string]int32{ "ten": 10, }`, types.StringType, types.Int32Type, "ten", "10"},
	}

	for i, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()
		requireNoParserErrors(t, p)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("test %d: program.Statements[0] not ast.ExpressionStatement. Got=%T", i, program.Statements[0])
		}

		mapExp, ok := stmt.Expression.(*ast.MapLiteral)
		if !ok {
			t.Fatalf("test %d: exp not *ast.MapLiteral. Got=%T", i, stmt.Expression)
		}

		mapTypes, ok := mapExp.Type.(*types.Map)
		if !ok {
			t.Fatalf("test %d: exp.Type not *types.Map. Got=%T", i, mapExp.Type)
		}

		if !types.IsTypesEqual(tt.expectedKeyType, mapTypes.Key) {
			t.Errorf(
				"test %d: expected key to be type: %s, got: %s", 
				i, 
				tt.expectedKeyType.Name(),
				mapTypes.Key.Name(),
			)
		}

		if !types.IsTypesEqual(tt.expectedValueType, mapTypes.Value) {
			t.Errorf(
				"test %d: expected value to be type: %s, got: %s", 
				i, 
				tt.expectedValueType.Name(),
				mapTypes.Value.Name(),
			)
		}

		if len(mapExp.Pairs) != 1 {
			t.Fatalf("test %d: expected 1 pairs, got %d", i, len(mapExp.Pairs))
		}

		pair := mapExp.Pairs[0]
		if pair.Key.String() != tt.expectedKey {
			t.Errorf(
				"test %d: expected key to be: %s, got: %s", 
				i, 
				tt.expectedKey,
				pair.Key.String(),
			)
		}

		if pair.Value.String() != tt.expectedValue {
			t.Errorf(
				"test %d: expected value to be: %s, got: %s", 
				i, 
				tt.expectedValue,
				pair.Value.String(),
			)
		}
	}
}

func TestNestedMapLiteral(t *testing.T) {
tests := []struct{
		input string
		outerLength int
		outerKey string
		innerLength int
		innerKey string
		innerValue string
	}{
		{
			`map[string]map[string]string{ 
	"a": map[string]string{
		"alan": "barrage",
		"andrew": "carryon",
	},
}`, 
			1,
			"a",
			2,
			"alan",
			"barrage",
		},
	}

	for i, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()
		requireNoParserErrors(t, p)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("test %d: program.Statements[0] not ast.ExpressionStatement. Got=%T", i, program.Statements[0])
		}

		mapExp, ok := stmt.Expression.(*ast.MapLiteral)
		if !ok {
			t.Fatalf("test %d: exp not *ast.MapLiteral. Got=%T", i, stmt.Expression)
		}

		if len(mapExp.Pairs) != tt.outerLength {
			t.Fatalf("test %d: expected %d pairs, got %d", i, tt.outerLength, len(mapExp.Pairs))
		}

		pair := mapExp.Pairs[0]
		if pair.Key.String() != tt.outerKey {
			t.Fatalf("test %d: expected outer key %s, got %s", i, tt.outerKey, pair.String())
		}

		value, ok := pair.Value.(*ast.MapLiteral)
		if !ok {
			t.Fatalf("test %d: expected value to be *ast.MapLiteral. Got=%T", i, pair.Value)
		}

		if len(value.Pairs) != tt.innerLength {
			t.Fatalf("test %d: expected inner length to have %d pairs, got %d", i, tt.innerLength, len(value.Pairs))
		}

		innerValue := value.Pairs[0]
		if innerValue.Key.String() != tt.innerKey {
			t.Fatalf(
				"test %d: expected inner key %s, got %s", 
				i, 
				tt.innerKey, 
				innerValue.Key.String(),
			)
		}

		if innerValue.Value.String() != tt.innerValue {
			t.Fatalf(
				"test %d: expected inner key %s, got %s", 
				i, 
				tt.innerValue, 
				innerValue.Value.String(),
			)
		}
	}
}

func TestComplicatedMapLiteral(t *testing.T) {
tests := []struct{
		input string
		outerLength int
		outerKey string
		innerLength int
		innerKeys []string
		innerValues []string
	}{
		{
			`struct Person {
	name: string;
	fn new(name string): Person {
		return Person{ name };
	}
	
	impl {
		fn getName(): string { return self.name; }
	}
}


a := Person.new("alan");
a2 := Person.new("andrew");

map[string]map[string]string{ 
	"a": map[string]string{
		a.getName(): a,
		a2.getName(): a2,
	},
}`, 
			1,
			"a",
			2,
			[]string{ "a.getName()", "a2.getName()"},
			[]string{ "a", "a2"},
		},
	}

	for i, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()
		requireNoParserErrors(t, p)

		stmt, ok := program.Statements[3].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("test %d: program.Statements[3] not ast.ExpressionStatement. Got=%T", i, program.Statements[0])
		}

		mapExp, ok := stmt.Expression.(*ast.MapLiteral)
		if !ok {
			t.Fatalf("test %d: exp not *ast.MapLiteral. Got=%T", i, stmt.Expression)
		}

		if len(mapExp.Pairs) != tt.outerLength {
			t.Fatalf("test %d: expected %d pairs, got %d", i, tt.outerLength, len(mapExp.Pairs))
		}

		pair := mapExp.Pairs[0]
		if pair.Key.String() != tt.outerKey {
			t.Fatalf("test %d: expected outer key %s, got %s", i, tt.outerKey, pair.String())
		}

		value, ok := pair.Value.(*ast.MapLiteral)
		if !ok {
			t.Fatalf("test %d: expected value to be *ast.MapLiteral. Got=%T", i, pair.Value)
		}

		if len(value.Pairs) != tt.innerLength {
			t.Fatalf("test %d: expected inner length to have %d pairs, got %d", i, tt.innerLength, len(value.Pairs))
		}

		for k, inner := range value.Pairs {
			keyCall, ok := inner.Key.(*ast.CallExpression)
			if !ok {
				t.Fatalf("expected key to be *ast.CallExpression, got %T", inner.Key)
			}

			_, ok = keyCall.Function.(*ast.MemberExpression)
			if !ok {
				t.Fatalf("expected call function to be *ast.MemberExpression, got %T", keyCall.Function)
			}

			if inner.Key.String() != tt.innerKeys[k] {
				t.Fatalf(
					"test %d, pair %d expected inner key %s, got %s", 
					i, 
					k,
					tt.innerKeys[k], 
					inner.Key.String(),
				)
			}

			iv, ok := inner.Value.(*ast.Identifier)
			if !ok {
				t.Fatalf(
					"test %d, pair %d: expected innerValue to be *ast.Identifier, got=%T",
					i,
					k,
					inner.Value,
				)
			}

			if iv.String() != tt.innerValues[k] {
				t.Fatalf(
					"test %d, pair %d expected inner value %s, got %s", 
					i, 
					k,
					tt.innerValues[k], 
					inner.Key.String(),
				)
			}
		}

		
	}
}

func TestMapDeclaration(t *testing.T) {
	tests := []struct{
		input string
		expectedIdentifier string
		expectedKeyType types.Type
		expectedValueType types.Type
		expectedKey string
		expectedValue string
	}{
		{`a: map[int32]string := map[int32]string{ 10: "ten", }`, "a", types.Int32Type, types.StringType, "10", "ten"},
	}

	for i, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()
		requireNoParserErrors(t, p)

		stmt, ok := program.Statements[0].(*ast.DeclareStatement)
		if !ok {
			t.Fatalf("test %d: program.Statements[0] not ast.DeclareStatement. Got=%T", i, program.Statements[0])
		}

		ident := stmt.Name.Value
		if ident != tt.expectedIdentifier {
			t.Fatalf("test %d: stmt.Name.Value not=%s. Got=%s", i, tt.expectedIdentifier, ident)
		}

		mapExp, ok := stmt.Value.(*ast.MapLiteral)
		if !ok {
			t.Fatalf("test %d: exp not *ast.MapLiteral. Got=%T", i, stmt.Value)
		}

		mapTypes, ok := mapExp.Type.(*types.Map)
		if !ok {
			t.Fatalf("test %d: exp.Type not *types.Map. Got=%T", i, mapExp.Type)
		}

		if !types.IsTypesEqual(tt.expectedKeyType, mapTypes.Key) {
			t.Errorf(
				"test %d: expected key to be type: %s, got: %s", 
				i, 
				tt.expectedKeyType.Name(),
				mapTypes.Key.Name(),
			)
		}

		if !types.IsTypesEqual(tt.expectedValueType, mapTypes.Value) {
			t.Errorf(
				"test %d: expected value to be type: %s, got: %s", 
				i, 
				tt.expectedValueType.Name(),
				mapTypes.Value.Name(),
			)
		}

		if len(mapExp.Pairs) != 1 {
			t.Fatalf("test %d: expected 1 pairs, got %d", i, len(mapExp.Pairs))
		}

		pair := mapExp.Pairs[0]
		if pair.Key.String() != tt.expectedKey {
			t.Errorf(
				"test %d: expected key to be: %s, got: %s", 
				i, 
				tt.expectedKey,
				pair.Key.String(),
			)
		}

		if pair.Value.String() != tt.expectedValue {
			t.Errorf(
				"test %d: expected value to be: %s, got: %s", 
				i, 
				tt.expectedValue,
				pair.Value.String(),
			)
		}
	}
}

func TestAccessMapValueByString(t *testing.T) {
	tests := []struct{
		input string
		access string
	}{
		{`map[string]string{ "one": "one", }["one"]`, "one"},
	}

	for i, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()
		requireNoParserErrors(t, p)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("test %d: program.Statements[0] not ast.ExpressionStatement. Got=%T", i, program.Statements[0])
		}

		indexExp, ok := stmt.Expression.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("test %d: program.Statements[0] not ast.IndexExpression. Got=%T", i, stmt.Expression)
		}

		if indexExp.Index.String() != tt.access {
			t.Fatalf("test %d: expected accesser to be '%s' got %s", i, tt.access, indexExp.Index.String())
		}
	}
}

func TestAssignToMap(t *testing.T) {
	tests := []struct{
		input string
		ident string
		indexValue string
		value string
	}{
		{`x := map[string]string{}; x["one"] = "one"`, "x", "one", "one"},
		{`x := map[string]string{ "one": "one", }; x["two"] = "two"`, "x", "two", "two"},
	}

	for i, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()
		requireNoParserErrors(t, p)

		as := program.Statements[1].(*ast.AssignStatement)

		index, ok := as.Target.(*ast.IndexExpression)
		if !ok {
			t.Fatalf(
				"test %d: expected as.Target to be ast.IndexExpression got %T",
				i,
				as.Target,
			)
		}

		if index.Left.String() != tt.ident {
			t.Errorf(
				"test %d: expected index.Left.String to be %s got %s",
				i,
				tt.ident,
				index.Left.String(),
			)
		}

		if index.Index.String() != tt.indexValue {
			t.Errorf(
				"test %d: expected index.Index.String to be %s got %s",
				i,
				tt.indexValue,
				index.Index.String(),
			)
		}
	}
}