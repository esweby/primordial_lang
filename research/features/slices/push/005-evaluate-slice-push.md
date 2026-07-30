# PUSH-005 — Evaluate slice `push`

## Description

Implement runtime method dispatch and fluent slice mutation after syntax,
semantic checks, and copy isolation are in place.

## Technical description

Extend call evaluation to recognize a member-expression callee. Evaluate the
receiver exactly once, then evaluate arguments from left to right. Dispatch the
`push` member to a slice-specific evaluator rather than registering a global
builtin.

The runtime implementation should append the argument to
`object.Slice.Elements` and return the same `*object.Slice`. Go may replace the
underlying element array during append; retaining the enclosing object preserves
the identity required by chaining.

Include defensive runtime checks for receiver kind and arity even though the
semantic analyser rejects invalid programs. Do not add a bound-method object in
this ticket because storing `slice.push` is outside the first-version contract.

## Acceptance criteria

- Calling `push` appends exactly one evaluated value.
- The receiver is evaluated once and before the argument.
- The returned object is the same runtime slice receiver.
- `slice.push(1).push(2)` mutates one receiver and evaluates successfully.
- Pushing to an empty slice works.
- Runtime arity and receiver guards return language errors rather than panicking.
- `push` is not exposed as a global builtin.
- The implementation does not add a bound-method runtime object.
- Existing global calls, slice literals, and indexing continue to evaluate
  correctly.
