# PUSH-006 — Add integration coverage and examples

## Description

Close the feature with end-to-end tests that prove syntax, static semantics,
runtime mutation, chaining, and copy isolation agree.

## Technical description

Add table-driven parser, semantic, and evaluator tests plus at least one full
program example. Cover successful pushes for empty and populated slices,
expression-valued chaining, and different supported element types.

Negative tests should cover arrays, non-collection receivers, unknown methods,
wrong arity, incompatible element types, malformed member syntax, and bare
method access. Copy tests must demonstrate that an assigned/declaration-copied
slice remains independent while every call within a chain observes prior calls.

Update user-facing collection documentation with the canonical syntax and one
brief note explaining that the binding need not be `mut` because `push` changes
slice contents rather than rebinding the variable.

## Acceptance criteria

- Parser tests cover one call, a chain, malformed dots, and composition with
  indexing.
- Semantic tests cover the returned `[]T` type and every documented rejection.
- Evaluator tests cover empty, populated, and chained pushes.
- At least two element types are exercised successfully.
- Tests prove receiver and argument evaluation occur once in the documented
  order.
- Tests prove declaration copies do not share later `push` mutations.
- An array `push` attempt is rejected before runtime.
- Documentation contains a runnable fluent example.
- The full Go test suite passes without weakening existing assertions.
