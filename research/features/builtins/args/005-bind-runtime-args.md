# ARG-005 — Bind `args()` into the runtime prelude

## Description

Implement the evaluator-facing `args()` builtin using the injected runtime context.

## Technical description

Bind the shared `args` definition to an `object.Builtin` closure that captures a copied argument snapshot. On each call:

1. defensively reject non-zero runtime arity;
2. allocate a new `[]object.Object`;
3. convert every host string to `*object.String`;
4. return `*object.Slice`.

Create a runtime prelude environment containing bound builtins, then create the user program environment as its child. Prefer this path over the evaluator's package-global fallback map. Existing builtins may be migrated incrementally, but the lookup order must remain explicit and tested.

## Acceptance criteria

- `args()` returns an `*object.Slice`.
- Every result element is an `*object.String` with the expected value.
- Argument order and empty strings are preserved.
- No supplied arguments produces a non-nil empty slice.
- Calling the runtime function with arguments produces a clear arity error.
- Two calls return slices with independent backing storage.
- Mutating one returned slice cannot affect the runtime context or a later call.
- Evaluator tests use injected arguments rather than modifying `os.Args`.
