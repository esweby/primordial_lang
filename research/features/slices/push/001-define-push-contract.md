# PUSH-001 — Define the fluent `push` contract

## Description

Settle the public behaviour of single-element slice growth before implementing
syntax or runtime dispatch. Reconcile the requested `push` name with the Design
Notes, which currently call the equivalent operation `append`.

## Technical description

Amend the Design Notes and language specification to define:

```primordial
slice.push(value T): []T
```

Specify that the method mutates a slice by appending one compatible value and
returns the same receiver. Plain bindings permit this interior mutation; `mut`
continues to govern rebinding. Arrays do not expose the method.

Replace the current single-element `.append(T)` entry with `.push(T)`. Reserve
`append` or `extend` for a future bulk operation rather than shipping duplicate
names. State that a future constant receiver cannot call mutating methods.

Document evaluation order, empty-slice behaviour, copy isolation, complexity,
and the decision not to support bare method values in the first release.

## Acceptance criteria

- The specification defines `push` for slices with exactly one argument.
- The result is explicitly the same mutated receiver with static type `[]T`.
- `push` is explicitly unavailable on arrays.
- Plain-binding interior mutation and `mut` rebinding are distinguished.
- The existing `.append(T)` wording is removed or redirected by an explicit
  naming decision.
- Receiver-before-argument evaluation order is documented.
- Deep-copy isolation across declarations is demonstrated by an example.
- Empty slices, constant receivers, and bare method values have defined
  behaviour or scope.
- Amortized O(1) and possible O(n) growth are documented.
