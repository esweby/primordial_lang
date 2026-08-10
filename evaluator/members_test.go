package evaluator

import (
	"math/big"
	"strings"
	"testing"

	"github.com/esweby/primordial_lang/object"
	"github.com/esweby/primordial_lang/types"
)

func TestMemberLength(t *testing.T) {
	tests := []struct {
		input  string
		output int64
	}{
		{`[]int64{20, 32, 45}.length;`, 3},
		{`[3]int64{20, 32, 45}.length;`, 3},
		{`[4]int64{20}.length;`, 4},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)

		res, ok := evaluated.(*object.Integer)
		if !ok {
			t.Fatalf("test %d failed: expected object.Integer but got %T", i, evaluated)
		}

		if res.Value.Cmp(big.NewInt(tt.output)) != 0 {
			t.Fatalf("test %d failed: expected %d but got %s", i, tt.output, res.Value.String())
		}
	}
}

func TestEvalStructTypeFunctionsAndMethods(t *testing.T) {
	evaluated := testEval(`
		struct Person {
			name: string;
			pub age: int32;
			active = true;

			fn new(name string, age int32): Person {
				return Person{name, age};
			}

			impl {
				fn getName(): string { return self.name; }
				fn birthday(): int32 {
					self.age = self.age + 1;
					return self.age;
				}
			}
		}

		person := Person.new("Ada", 41);
		person.birthday();
		person.age;
	`)

	integer, ok := evaluated.(*object.Integer)
	if !ok {
		t.Fatalf("expected int32 age, got %T (%v)", evaluated, evaluated)
	}
	if integer.Value.Cmp(big.NewInt(42)) != 0 {
		t.Fatalf("expected age 42, got %s", integer.Value.String())
	}
	if !types.IsTypesEqual(integer.IntegerType, types.Int32Type) {
		t.Fatalf("expected runtime age type int32, got %s", integer.IntegerType.Name())
	}
}

func TestEvalStructMethodReadsPrivateFields(t *testing.T) {
	evaluated := testEval(`
		struct Person {
			name: string;
			fn new(name string): Person { return Person{name}; }
			impl { fn getOtherName(other Person): string { return other.name; } }
		}
		ada := Person.new("Ada");
		grace := Person.new("Grace");
		ada.getOtherName(grace);
	`)

	value, ok := evaluated.(*object.String)
	if !ok || value.Value != "Grace" {
		t.Fatalf("expected Grace, got %T (%v)", evaluated, evaluated)
	}
}

func TestEvalNestedStructMethodMutation(t *testing.T) {
	evaluated := testEval(`
		struct Address {
			pub city: string;
			impl { fn setCity(city string) { self.city = city; } }
		}
		struct Person { pub address: Address; }
		person := Person{address: Address{city: "London"}};
		person.address.setCity("Paris");
		person.address.city;
	`)

	value, ok := evaluated.(*object.String)
	if !ok || value.Value != "Paris" {
		t.Fatalf("expected Paris, got %T (%v)", evaluated, evaluated)
	}
}

func TestEvalStructMethodMayAssignItsField(t *testing.T) {
	evaluated := testEval(`
		struct Person {
			pub name: string;
			fn new(): Person { return Person{name: ""}; }
			impl {
				fn setName(newName string) { self.name = newName; }
			}
		}
		person := Person.new();
		person.setName("Ada");
		person.name;
	`)

	value, ok := evaluated.(*object.String)
	if !ok || value.Value != "Ada" {
		t.Fatalf("expected Ada, got %T (%v)", evaluated, evaluated)
	}
}

func TestEvalRejectsInvalidStructFieldAccess(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "private read",
			input:    `struct Person { name: string; } person := Person{name: "Ada"}; person.name;`,
			expected: "field Person.name is private",
		},
		{
			name:     "external write",
			input:    `struct Person { pub name: string; } person := Person{name: "Ada"}; person.name = "Grace";`,
			expected: "cannot assign to field outside its struct: name",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			evaluated := testEval(test.input)
			err, ok := evaluated.(*object.Error)
			if !ok || !strings.Contains(err.Message, test.expected) {
				t.Fatalf("expected error containing %q, got %T (%v)", test.expected, evaluated, evaluated)
			}
		})
	}
}

func TestMemberSliceAppend(t *testing.T) {
	tests := []struct {
		input  string
		output []int64
	}{
		{`ages := []int64{20, 32, 45}; ages.append(19); ages`, []int64{20, 32, 45, 19}},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)

		res, ok := evaluated.(*object.Slice)
		if !ok {
			t.Fatalf("test %d failed: expected object.Slice but got %T", i, evaluated)
		}

		for k, cur := range tt.output {
			testIntegerObject(t, res.Elements[k], cur)
		}
	}
}
