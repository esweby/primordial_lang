# ARG-007 — Add end-to-end coverage and user documentation

## Description

Close the feature with tests across the semantic, runtime, launcher, and user-facing boundaries.

## Technical description

Add focused unit tests for the builtin checker and runtime binder, plus integration tests that invoke the runner with controlled argument slices. Avoid tests that mutate the test binary's real `os.Args`.

Add a small `.pri` fixture that calls `args()` and exercises `len` and indexing once those operations are semantically available. Document the final invocation syntax in the README and language specification.

Include a regression test demonstrating that semantic analysis and evaluation use the same builtin registry; a builtin present in only one phase should make registry construction or test setup fail.

## Acceptance criteria

- Tests cover zero, one, and multiple program arguments.
- Tests preserve an empty string, a string containing spaces, and a leading-dash argument.
- Tests prove `args()` has semantic type `[]string`.
- Tests prove `args(1)` fails semantically and defensively at runtime.
- Tests prove returned slices are independent between calls.
- An integration test proves launcher tokens are stripped.
- An integration test covers the optional `--` separator.
- A regression test detects semantic/runtime builtin registry drift.
- README usage documentation includes a complete `primordial run` example.
- `go test ./...` passes.
