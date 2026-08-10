# Primordial documentation

This directory describes the implemented language contracts that are useful to
people working on Primordial itself. The language specification remains the
design document; these notes explain how the current implementation behaves.

## Integers

Primordial has eight fixed-width integer types:

| Type | Width | Minimum | Maximum |
| --- | ---: | ---: | ---: |
| `int8` | 8 | -128 | 127 |
| `uint8` | 8 | 0 | 255 |
| `int16` | 16 | -32,768 | 32,767 |
| `uint16` | 16 | 0 | 65,535 |
| `int32` | 32 | -2,147,483,648 | 2,147,483,647 |
| `uint32` | 32 | 0 | 4,294,967,295 |
| `int64` | 64 | -9,223,372,036,854,775,808 | 9,223,372,036,854,775,807 |
| `uint64` | 64 | 0 | 18,446,744,073,709,551,615 |

`int` and `uint` appear in the design specification but are not implemented
yet. Their target-dependent width needs a separate language decision. An
integer literal without another type context currently defaults to `int64`.

### Integer literals

An integer literal starts as an exact, untyped integer constant. Parsing does
not force it into a signed 64-bit host value, so the full `uint64` range and the
minimum `int64` value are accepted:

```pri
minimum: int64 := -9223372036854775808;
maximum: uint64 := 18446744073709551615;
```

The semantic analyser gives the literal a concrete type using its context:

```pri
small: int8 := 12;          // 12 becomes int8
items := []uint16{1, 2, 3}; // each literal becomes uint16

fn identity(value int32): int32 {
    return value;
}

result := identity(42);     // 42 becomes int32
```

This contextual materialisation is used by declarations, assignments,
function arguments and returns, collection elements, struct fields, member
arguments, integer operators, indexes, and value-producing `if` branches.

If no context supplies a type, the completed constant expression becomes
`int64`:

```pri
value := 1 + 2; // int64
```

### Representability

An untyped constant can become a concrete integer only when its mathematical
value fits that type. There is no silent truncation and negative constants
cannot become unsigned integers:

```pri
ok: int8 := 127;
alsoOk: uint8 := 255;

tooLarge: int8 := 128; // semantic error
negative: uint8 := -1; // semantic error
```

Constant expressions are checked using their final exact value:

```pri
ok: int8 := 100 + 27;
overflow: int8 := 100 + 28; // semantic error
```

### Operators and type inference

An untyped integer operand adopts the type of the concrete integer operand on
the other side. Operand order therefore does not affect the result type:

```pri
x: int32 := 10;
a := x + 2; // int32
b := 2 + x; // int32
c := x < 20; // valid comparison
```

Two operands that already have concrete but different integer types are not
implicitly promoted:

```pri
x: int32 := 1;
y: int64 := 2;
invalid := x + y; // semantic error
```

This rule keeps conversions visible and prevents operand order from selecting
a result type. Equality and ordering comparisons follow the same operand
resolution rules.

Unary negation is available for signed integers and untyped constants. It is
not available for an unsigned runtime value:

```pri
count: uint32 := 1;
invalid := -count; // semantic error
```

### Explicit integer conversions

Call-like type syntax performs a checked conversion:

```pri
small: int32 := 12;
wide := int64(small); // int64
byte := uint8(wide);  // uint8, provided the runtime value fits
```

Conversions never truncate or wrap. A constant conversion that cannot fit is
rejected during semantic analysis. A conversion of a runtime value is checked
by the evaluator and returns a Primordial error if it is out of range.

### Overflow

Ordinary arithmetic is checked. Addition, subtraction, multiplication,
division, and negation must produce a value representable by the result type:

```pri
maximum: int8 := 127;
maximum + 1; // ERROR: integer overflow
```

Division by zero is rejected. The signed minimum divided by `-1` is also an
overflow because the positive result does not fit the original signed type.
Primordial does not currently provide wrapping or saturating operators.

### Runtime representation

The parser and evaluator use Go's `math/big.Int` for exact integer values. A
runtime integer object also carries its concrete Primordial type, preserving
both width and signedness:

```text
IntegerObject
├── exact mathematical value
└── Primordial integer type
    ├── bit width
    └── signed or unsigned
```

Using arbitrary precision in the tree-walking evaluator prioritises a clear
reference implementation over interpreter speed. Values are still constrained
to their declared Primordial range whenever they are materialised or produced
by an operation.

The future LLVM backend will lower the type width to `iN`. Primordial's
signedness metadata will select signed or unsigned division, comparisons, and
extensions during lowering.

### Arrays, indexes, and lengths

Partially initialised integer arrays receive zero values carrying the array's
element type. A `[4]uint16{1}` therefore contains four `uint16` values, not
generic integers.

Indexes accept concrete integer values. Untyped index constants default to
`int64`. Negative and out-of-bounds indexes produce evaluator errors.

Collection `length` and the `len` builtin currently return `int64`. This can be
revisited together with the future definition of target-sized `int` and
`uint`.
