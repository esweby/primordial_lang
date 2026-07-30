# ARG-004 — Integrate builtins into semantic analysis

## Description

Make builtin calls visible to the semantic analyser so `args()` is checked and inferred exactly like other callable values.

## Technical description

Create a semantic prelude from the shared builtin registry. Represent each definition as a `BuiltinSymbol` carrying its call checker. Place the user program's symbol table in an enclosed scope above the prelude so lookup finds builtins while normal lexical shadowing remains possible.

Update `analyzeCallExpression` to recognize builtin symbols, analyse every supplied argument expression, invoke the definition's checker, and return its result type. Do not special-case the name `"args"` in the expression analyser.

This ticket should also stop treating all non-`FunctionSymbol` function values as having an unknowable generic result when a concrete `types.Function` signature is available.

## Acceptance criteria

- A freshly configured analyser resolves `args` without a user declaration.
- `args()` analyses to `[]string`.
- `args(1)` produces exactly one clear arity diagnostic.
- `arguments := args()` registers `arguments` with type `[]string`.
- A user binding in the program scope follows the documented builtin-shadowing rule.
- Builtins are not inserted into every nested scope.
- Runtime objects are not required to perform semantic analysis.
- Existing named-function and collection semantic tests continue to pass.
