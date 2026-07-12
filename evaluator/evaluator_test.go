package evaluator

import (
	"testing"

	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/object"
	"github.com/esweby/primordial_lang/parser"
	"github.com/esweby/primordial_lang/semantic"
)

func TestEvalIntegerExpr(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 - 50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpr(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true}, //0
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"2 > 1", true},
		{"1 == 1", true}, // 7
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"true == false", false},
		{"false == false", true},
		{"false == true", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 < 2) != true", false},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, i, evaluated, tt.expected)
	}
}

func TestDeclareStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`a := 5; a;`, 5},
		{`a := 5 * 5; a;`, 25},
		{`a := 5; b := a; b;`, 5},
		{`a := 5; b := 5; c := a * b; c;`, 25},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestIfElseExpr(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 > 2) { 10 } else if(2 > 1) { 20 }", 20},
		{"if (1 > 2) { 10 } else if(2 == 1) { 20 } else { 30 }", 30},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestReturnIntStatement(t *testing.T) {
	tests := []struct {
		input      string
		numReturns int
		expected   int64
	}{
		{"return 10;", 1, 10},
		{"return 10; 9;", 1, 10},
		{"9; return 10;", 1, 10},
		{"9; return 10 * 2; 10", 1, 20},
		{`if (10 > 1) { if (10 > 1) { return 1; }}`, 1, 1},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)
		ro := evaluated.(*object.ReturnValue)

		if len(ro.Value) != tt.numReturns {
			t.Fatalf("test %d: return object does not have %d return values. Got=%d", i, tt.numReturns, len(ro.Value))
		}

		if len(ro.Value) == 1 {
			testIntegerObject(t, ro.Value[0], tt.expected)
		}
	}
}

func TestFunctionLiteral(t *testing.T) {
	tests := []struct {
		input     string
		numParams int
		numReturn int
	}{
		{`(fn(x int32, y int32): int32 { return x + y; });`, 2, 1},
		{`add := fn(x int32, y int32): int32 { return x + y; }; add;`, 2, 1},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)
		testFunction(t, evaluated, i, tt.numParams, tt.numReturn)
	}
}

func TestFunctionStatement(t *testing.T) {
	tests := []struct {
		input     string
		numParams int
		numReturn int
	}{
		{`fn add(): int32 { return 3 + 3; }; add`, 0, 1},
		{`fn add(x int32, y int32): int32 { return x + y; }; add;`, 2, 1},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)
		testFunction(t, evaluated, i, tt.numParams, tt.numReturn)
	}
}

func TestFunctionCall(t *testing.T) {
	tests := []struct {
		input  string
		output int64
	}{
		{`identity := fn(): int64 { return 4000; } identity()`, 4000},
		{`fn add(x int32, y int32): int64 { return x + y; }; add(5, 5);`, 10},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.output)
	}
}

func TestFunctionTupleReturn(t *testing.T) {
	tests := []struct {
		input        string
		tupleLen     int
		firstReturn  int
		secondReturn bool
	}{
		{
			`
				fn values(): int32, bool {
					return 10, true;
				};

				values();
			`,
			2,
			10,
			true,
		},
		{
			`
				fn values(): int32, bool {
				if (true) {
					return 10, true;
				}

				return 20, false;
			}

			values();
			`,
			2,
			10,
			true,
		},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)
		tuple, ok := evaluated.(*object.Tuple)
		if !ok {
			t.Fatalf(
				"test %d: expected object.Tuple, got=%T (%+v)",
				i,
				evaluated,
				evaluated,
			)
		}

		if len(tuple.Elements) != 2 {
			t.Fatalf(
				"test %d: expected 2 tuple elements, got=%d",
				i, len(tuple.Elements),
			)
		}

		testIntegerObject(t, tuple.Elements[0], 10)
		testBooleanObject(t, 0, tuple.Elements[1], true)
	}
}

func TestFunctionClosures(t *testing.T) {
	tests := []struct{
		input string
		output int64
	}{
		{`
			newAdder := fn(x int64): function {
				return fn(y int64): int64 {
					return x + y;
				};
			};
			addFive := newAdder(5);
			addFive(5);
		`, 10},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.output)
	}
}

func TestTupleDeclaration(t *testing.T) {
	evaluated := testEval(`
		fn values(): int32, bool { return 10, true; };
		(number, _) := values();
		number;
	`)

	testIntegerObject(t, evaluated, 10)
}

func TestTupleAssignment(t *testing.T) {
	evaluated := testEval(`
		fn values(): int32, int32 { return 10, 20; };
		mut first := 0;
		mut second := 0;
		(first, second) = values();
		first + second;
	`)

	testIntegerObject(t, evaluated, 30)
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{`5 + true;`, "type mismatch: INTEGER + BOOLEAN"},
		{`5 + true; 5;`, "type mismatch: INTEGER + BOOLEAN"},
		{"-true", "unknown operator: -BOOLEAN"},
		{"true + false", "unknown operator: BOOLEAN + BOOLEAN"},
		{"5; true + false; 5", "unknown operator: BOOLEAN + BOOLEAN"},
		{"foobar;", "identifier not found: foobar"},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)

		if !ok {
			t.Errorf("test %d: no error object returned. Got=%T(%+v)", i, evaluated, evaluated)
			continue
		}

		if errObj.Message != tt.message {
			t.Errorf("test %d: wrong error message. expected=%s. got=%s", i, tt.message, errObj.Message)
		}
	}
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!!true", true},
		{"!!false", false},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, i, evaluated, tt.expected)
	}
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	symbols := semantic.NewSymbolTable()
	sa := semantic.NewSemanticAnalyzer(program, symbols)
	sa.Analyze()
	env := object.NewEnvironment()
	return Eval(program, env)
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
		return false
	}

	return true
}

func testBooleanObject(t *testing.T, i int, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("test %d: object has wrong value. got=%t, want=%t", i, result.Value, expected)
		return false
	}

	return true
}

func testFunction(t *testing.T, fn object.Object, testNum, numParams, numReturns int) {
	f, ok := fn.(*object.Function)
	if !ok {
		t.Fatalf("test %d: test is not object.Function. Got=%T (%+v)", testNum, fn, fn)
	}

	if len(f.Parameters) != numParams {
		t.Fatalf("test %d: incorrect num params. Got=%d. Want=%d", testNum, len(f.Parameters), numParams)
	}

	if len(f.ReturnTypes) != numReturns {
		t.Fatalf("test %d: incorrect num params. Got=%d. Want=%d", testNum, len(f.ReturnTypes), numReturns)
	}
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != nil {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}

	return true
}
