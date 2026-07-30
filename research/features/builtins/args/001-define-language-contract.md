# ARG-001 — Define the `args()` language contract

## Description

Document the public behaviour of `args()` before implementing its runtime representation. The contract must settle which host arguments are visible, how the REPL behaves, and whether calls return shared or independent values.

## Technical description

Add an `args()` section to the language specification with the signature:

```primordial
args(): []string
```

Define the result as user-program arguments only. Exclude the Primordial executable, `run`, the source path, and the optional `--` separator. Specify that the result is ordered, contains a distinct string for each host argument, is empty in the REPL, and is freshly allocated for each invocation.

Document UTF-8 validation as a launcher responsibility. `args()` itself should be infallible once program execution begins.

## Acceptance criteria

- The specification declares `args()` as a zero-argument builtin returning `[]string`.
- The first returned element is defined as the first user-supplied program argument.
- Executable, subcommand, source path, and delimiter tokens are explicitly excluded.
- Empty invocation returns an empty slice rather than an error or nil value.
- Each call is specified to return an independently mutable slice.
- REPL behaviour is explicitly defined as returning an empty slice.
- Invalid UTF-8 behaviour is defined as a launcher diagnostic before execution.
- The specification includes at least one invocation and expected-result example.
