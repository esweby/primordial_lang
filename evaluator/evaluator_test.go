package evaluator

import (
	"math/big"
	"testing"

	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/object"
	"github.com/esweby/primordial_lang/parser"
	"github.com/esweby/primordial_lang/semantic"
)

func TestFunctionCall(t *testing.T) {
	tests := []struct {
		input  string
		output int64
	}{
		{`identity := fn(): int64 { return 4000; } identity()`, 4000},
		{`fn add(x int32, y int32): int32 { return x + y; }; add(5, 5);`, 10},
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
	tests := []struct {
		input  string
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

func TestStringLiteral(t *testing.T) {
	input := `"Hello world";`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("expected string object. Got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello world" {
		t.Fatalf("expected string value to be Hello world. Got=%s", str.Value)
	}
}

func TestStringConcatonation(t *testing.T) {
	input := `"Hello" + " " + "world";`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("expected string object. Got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello world" {
		t.Fatalf("expected string value to be Hello world. Got=%s", str.Value)
	}
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

func TestForLoops(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{} // int64, bool, string, or nil for void
	}{
		{
			name: "infinite loop with break",
			input: `
				x := 0;
				for {
					x = x + 1;
					if (x == 5) { break; }
				}
				x;
			`,
			expected: int64(5),
		},
		{
			name: "while loop",
			input: `
				x := 0;
				for (x < 5) {
					x = x + 1;
				}
				x;
			`,
			expected: int64(5),
		},
		{
			name: "constructed loop",
			input: `
				sum := 0;
				for (i := 1; i <= 5; i = i + 1) {
					sum = sum + i;
				}
				sum;
			`,
			expected: int64(15),
		},
		{
			name: "range over array",
			input: `
				arr := [3]int64{10, 20, 30};
				sum := 0;
				for i, val := range arr {
					sum = sum + val;
				}
				sum;
			`,
			expected: int64(60),
		},
		{
			name: "range over slice",
			input: `
				slice := []int64{1, 2, 3, 4};
				sum := 0;
				for _, val := range slice {
					sum = sum + val;
				}
				sum;
			`,
			expected: int64(10),
		},
		// {
		// 	name: "range over map",
		// 	input: `
		// 		m := map[string]int64{"a": 1, "b": 2, "c": 3};
		// 		sum := 0;
		// 		for _, val := range m {
		// 			sum = sum + val;
		// 		}
		// 		sum;
		// 	`,
		// 	expected: int64(6),
		// },
		{
			name: "break with label",
			input: `
				x := 0;
				outer: for {
					x = x + 1;
					for {
						break outer;
					}
				}
				x;
			`,
			expected: int64(1),
		},
		{
			name: "continue in constructed loop",
			input: `
				sum := 0;
				for (i := 1; i <= 5; i = i + 1) {
					if (i == 3) { continue; }
					sum = sum + i;
				}
				sum;
			`,
			expected: int64(12), // 1+2+4+5
		},
		{
			name: "loop as expression (break with value)",
			input: `
				x := for {
					break (42);
				};
				x;
			`,
			expected: int64(42),
		},
		{
			name: "loop as expression with break label and value",
			input: `
				x := lbl: for {
					break lbl (99);
				};
				x;
			`,
			expected: int64(99),
		},
		{
			name: "nested loops - break inner only",
			input: `
				outer := 0;
				inner := 0;
				for {
					outer = outer + 1;
					if (outer > 3) { break; }
					for {
						inner = inner + 1;
						if (inner > 2) { break; }
					}
				}
				outer + inner;
			`,
			expected: int64(9), // outer=4, inner=5 => 9
		},
		{
			name: "labelled break to outer",
			input: `
				outer := 0;
				inner := 0;
				outerLabel: for {
					outer = outer + 1;
					for {
						inner = inner + 1;
						if (inner > 2) { break outerLabel; }
					}
				}
				outer + inner;
			`,
			expected: int64(4), // outer=1, inner=3? Actually: inner becomes 1,2 then break outer, so outer=1, inner=3 => 4
		},
		{
			name: "loop returns void (no break value) - assignment allowed?",
			input: `
				x := for { break; };
				x;
			`,
			expected: nil, // void object
		},
		{
			name: "break outside loop should error",
			input: `
				break;
			`,
			expected: "break outside of loop", // error message string
		},
		{
			name: "continue outside loop should error",
			input: `
				continue;
			`,
			expected: "continue outside of loop",
		},
		{
			name: "break with label not matching any loop should error",
			input: `
				for { break unknownLabel; }
			`,
			expected: "break label 'unknownLabel' does not match any enclosing loop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)

			switch exp := tt.expected.(type) {
			case int64:
				testIntegerObject(t, evaluated, exp)
			case nil:
				// expected void – check that it's not an error and type is void?
				if evaluated == nil || evaluated == object.VOID {
					// ok
				} else {
					t.Errorf("expected void, got %T (%+v)", evaluated, evaluated)
				}
			case string:
				// expect error with message containing exp
				errObj, ok := evaluated.(*object.Error)
				if !ok {
					t.Errorf("expected error, got %T (%+v)", evaluated, evaluated)
					return
				}
				if errObj.Message != exp {
					t.Errorf("expected error message %q, got %q", exp, errObj.Message)
				}
			default:
				t.Fatalf("unsupported expected type: %T", exp)
			}
		})
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

	evaluator := New(program, env)

	return evaluator.Evaluate()
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value.Cmp(big.NewInt(expected)) != 0 {
		t.Errorf("object has wrong value. got=%s, want=%d", result.Value.String(), expected)
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

func testErrorObject(t *testing.T, obj object.Object) bool {
	_, ok := obj.(*object.Error)
	if !ok {
		t.Errorf("object is not Error. got=%T (%+v)", obj, obj)
		return false
	}

	return true
}

func hashKeyFromAny(t *testing.T, v any) object.HashKey {
	t.Helper()
	switch val := v.(type) {
	case int:
		return (&object.Integer{Value: big.NewInt(int64(val))}).HashKey()
	case int64:
		return (&object.Integer{Value: big.NewInt(val)}).HashKey()
	case string:
		return (&object.String{Value: val}).HashKey()
	case bool:
		return (&object.Boolean{Value: val}).HashKey()
	default:
		t.Fatalf("unsupported expected key type %T", v)
		return object.HashKey{}
	}
}
