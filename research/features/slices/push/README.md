# Fluent slice `push` research and delivery plan

Status: proposed

Target signature:

```primordial
slice.push(value T): []T
```

## Executive recommendation

Implement `push` as a mutating slice method that appends one value and returns the
same slice receiver. This gives Primordial the fluent form:

```primordial
values := []int32{1};
values.push(2).push(3);
```

After the expression, `values` contains `{1, 2, 3}`. `push` should work on a
normally declared slice because the Design Notes distinguish immutable bindings
from mutable collection contents. It must not work on arrays, whose length is
fixed.

Adopt `push(T)` as the canonical name for adding one item and amend the existing
Design Notes entry that calls this operation `append(T)`. Reserve `append` or
`extend` for a future bulk operation. Supporting both names immediately would
create an alias without adding capability.

## Fit with the Design Notes

The current [Design Notes](../../../../specification/Design%20Notes.md) establish
four relevant rules:

- slice contents are mutable through methods even when their binding is not;
- slices are variable-sized while arrays reject length-changing methods;
- values have neutral defaults rather than nil;
- assigning a collection to another declaration performs a deep copy.

The proposed method follows the first two rules directly. The last rule is an
important prerequisite: returning the receiver makes chaining straightforward,
but a declaration copy must not accidentally share that receiver.

```primordial
original := []int32{1};
copy := original;
original.push(2);

// original is {1, 2}; copy remains {1}
```

`mut` should continue to mean that a binding can be rebound. It should not be
required for interior slice mutation. A future `const` implementation should
reject `push` on a constant receiver; dynamic slices may instead be prohibited
as constant values.

## Research findings

Collection APIs make different return-value choices:

| Language | Operation | Mutation and result |
|---|---|---|
| Rust | `Vec::push(value)` | Mutates through `&mut self` and returns `()` |
| Swift | `Array.append(value)` | Mutates the array and returns `Void` |
| JavaScript | `Array.prototype.push(...items)` | Mutates and returns the new length |
| Go | `append(slice, values...)` | Returns a slice value that callers normally assign because storage may change |

See the official [Rust `Vec` documentation](https://doc.rust-lang.org/std/vec/struct.Vec.html#method.push),
[Swift `Array.append` documentation](https://developer.apple.com/documentation/swift/array/append%28_%3A%29),
[ECMAScript `Array.prototype.push`](https://tc39.es/ecma262/multipage/indexed-collections.html#sec-array.prototype.push),
and the [Go specification for `append`](https://go.dev/ref/spec#Appending_and_copying_slices).

None combines in-place mutation with receiver chaining. Primordial can do so
deliberately: the result is the receiver, not a new length, `void`, or a copied
slice. This is consistent with the requested fluent syntax and keeps the
operation distinguishable from Go's builtin `append`.

Like the established implementations, `push` is amortized O(1) and may be O(n)
when the backing storage grows. A reallocation changes the `Elements` storage,
but not the enclosing runtime slice object's identity.

## Proposed language contract

1. `push` is a method on `[]T`, not a global builtin.
2. It accepts exactly one expression assignable to `T`.
3. It evaluates the receiver once, then the argument once, from left to right.
4. It appends the evaluated value to the receiver.
5. It returns that same receiver with static type `[]T`.
6. It works on empty and non-empty slices.
7. It is unavailable on arrays and all non-slice receivers.
8. A plain binding may call it; rebinding permission is unrelated to content
   mutation.
9. A declaration copy remains isolated according to Primordial's deep-copy rule.
10. Bare method values such as `callback := values.push` are out of scope for
    the first implementation.

No zero-fill behaviour is involved. Unlike an array literal with a declared
length, a slice grows by exactly one element per successful call.

## Parsing and AST

Add `.` as a token and introduce a reusable member-access expression:

```text
CallExpression
├── Function: MemberExpression
│   ├── Left: Identifier("values")
│   └── Name: Identifier("push")
└── Arguments: [IntegerLiteral(2)]
```

Register member access as a high-precedence Pratt infix/postfix operation. The
existing call parser can then wrap the member expression. Repeated postfix
parsing naturally produces:

```primordial
values.push(2).push(3)
```

This AST is preferable to a `PushExpression`: it creates the syntax foundation
for the other slice methods already listed in the Design Notes.

## Semantic analysis

Extend call analysis to accept a member expression as the callee while retaining
the existing identifier-call path. The first method resolver can be intentionally
small:

- analyse the receiver and require `*types.Slice`;
- resolve only `push`;
- require one argument;
- check that the argument type is assignable to the slice element type;
- return the receiver's slice type.

Expose the slice element type through a method on `types.Slice` rather than
reaching into its private field. Reject unknown methods, wrong arity, incompatible
elements, array receivers, and bare method values with focused diagnostics.

## Runtime evaluation

For a member call, evaluate the receiver exactly once before evaluating its
arguments. Dispatch `push` through a slice-method evaluator, append to
`object.Slice.Elements`, and return the same `*object.Slice`.

A general bound-method runtime object is unnecessary for the first version
because method values are out of scope. Direct member-call dispatch is smaller
while leaving the AST general enough to add bound methods later.

Before relying on in-place mutation, enforce the Design Notes' copy boundary at
declaration and assignment. During a chain the evaluator preserves identity;
across a language-level copy it must create independent slice storage.

## Rejected alternatives

### Global `push(slice, value)`

This avoids member syntax but conflicts with the desired fluent API and does not
advance the method infrastructure needed by `popLast`, `splice`, and future
collection methods.

### Return the new length

JavaScript uses this result, but it prevents receiver chaining and makes the
expression's type unrelated to its receiver.

### Return no value

Rust and Swift make mutation statement-oriented. That is coherent, but it cannot
support the requested chained form.

### Return a fresh slice on every call

This would make `values.push(2)` ineffective unless assigned and would contradict
the Design Notes' mutable-content rule. Copies belong at Primordial assignment
boundaries, not at every method invocation.

## Delivery order

| Order | Ticket | Outcome |
|---:|---|---|
| 1 | [PUSH-001](001-define-push-contract.md) | Stable naming, mutation, and return contract |
| 2 | [PUSH-002](002-add-member-expression-syntax.md) | `receiver.method(...)` parses and chains |
| 3 | [PUSH-003](003-analyse-slice-push.md) | Invalid receivers, arguments, and calls are rejected |
| 4 | [PUSH-004](004-enforce-slice-copy-semantics.md) | Mutation cannot leak across declaration copies |
| 5 | [PUSH-005](005-evaluate-slice-push.md) | Runtime mutation and fluent return behaviour |
| 6 | [PUSH-006](006-add-integration-coverage.md) | End-to-end tests and user-facing examples |

PUSH-003 depends on PUSH-002. PUSH-004 can proceed after PUSH-001 and should be
complete before PUSH-005 is accepted. PUSH-006 closes the feature.

## Definition of done

The feature is complete when this program parses, passes semantic analysis, and
evaluates with `numbers` equal to `{1, 2, 3}`:

```primordial
numbers := []int32{1};
numbers.push(2).push(3);
```

Tests must also prove that arrays reject `push`, incompatible element types are
diagnosed, a copied slice is isolated, and the value of a `push` expression is
the mutated receiver.
