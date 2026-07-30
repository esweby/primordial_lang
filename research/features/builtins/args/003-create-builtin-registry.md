# ARG-003 — Create a shared builtin registry

## Description

Create one source of truth for builtin names, semantic call checking, and runtime binding.

## Technical description

Introduce a neutral builtin-definition package consumed by both semantic setup and runtime setup. Each definition should provide:

```go
type Definition struct {
    Name      string
    CheckCall func([]types.Type) (types.Type, error)
    Bind      func(RuntimeContext) *object.Builtin
}
```

Register `args` with a checker that accepts zero arguments and returns `types.NewSlice(types.StringType)`. Its runtime binder should be declared but may be completed in ARG-005.

The checker callback intentionally supports more than fixed signatures so polymorphic builtins such as `len` can eventually share the same registry. Migrate `len` in this ticket if doing so remains small; otherwise add a clearly tracked compatibility path rather than creating a second permanent registry.

Reject duplicate builtin names during registry construction.

## Acceptance criteria

- A builtin definition represents both semantic call checking and runtime binding.
- `args` has exactly one canonical definition.
- The `args` checker rejects one or more arguments.
- The `args` checker returns `[]string` for a zero-argument call.
- Registry construction rejects duplicate names.
- Callers can enumerate definitions deterministically.
- The package imports neither `semantic` nor `evaluator`.
- Tests cover lookup, duplicate rejection, valid `args()` checking, and invalid arity.
