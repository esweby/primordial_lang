# PUSH-003 — Analyse slice `push` calls

## Description

Add static checking for slice method calls so invalid receivers and values fail
before evaluation.

## Technical description

Extend call-expression analysis to handle a member expression as the callee
without changing the existing identifier-call path. Add an element-type accessor
to `types.Slice`, then introduce a focused slice-method resolver.

For `receiver.push(argument)`, analyse the receiver once, require a slice type,
require exactly one argument, and use the language's normal assignability check
against the slice element type. Return the receiver slice type so a second
`.push(...)` can be analysed in the same expression.

Reject `push` on arrays and other values, unknown slice methods, wrong arity, and
incompatible arguments with diagnostics that include the method and relevant
types. Reject a member expression used outside a direct call until bound method
values are designed.

## Acceptance criteria

- `types.Slice` exposes its element type without exporting its internal field.
- A valid `[]T.push(T)` expression analyses to `[]T`.
- Chained pushes pass the result type into the next call.
- Zero or multiple arguments produce an arity diagnostic.
- An incompatible element produces an assignability diagnostic.
- Arrays and non-slice receivers produce a method-not-available diagnostic.
- Unknown slice method names are rejected clearly.
- Bare method values are rejected explicitly.
- Existing builtin and user-function call analysis remains covered and passing.
