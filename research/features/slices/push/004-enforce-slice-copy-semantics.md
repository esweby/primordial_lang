# PUSH-004 — Enforce slice copy semantics

## Description

Protect Primordial's deep-copy-on-assignment rule before introducing visible
in-place slice mutation.

## Technical description

Audit declaration and assignment evaluation paths for collection values. When a
slice value crosses a language-level copy boundary, create an independent
`object.Slice` with independent element storage. Apply the same rule consistently
wherever the specification says a copy occurs.

Define whether copying is recursively deep for nested slices and collection
elements, then implement that exact rule through one reusable object-copy helper.
Do not copy a receiver between consecutive calls in a chain: `push` must return
and continue mutating the same runtime slice object.

Keep ordinary expression evaluation from cloning indiscriminately. Copying at
every identifier read would hide receiver mutation, waste allocations, and make
method behaviour dependent on incidental evaluation.

## Acceptance criteria

- Declaring one slice from another produces independent element storage.
- Pushing onto the source after a declaration copy does not alter the copy.
- Pushing onto the copy does not alter the source.
- The chosen nested-collection copy rule is documented and tested.
- One reusable helper owns the runtime object-copy behaviour.
- Literal initialization does not acquire unintended shared mutable storage.
- A chained call retains one receiver identity across every `push`.
- Tests distinguish language copy boundaries from ordinary identifier lookup.
- Existing scalar and array copy behaviour remains passing or is updated to the
  documented rule.
