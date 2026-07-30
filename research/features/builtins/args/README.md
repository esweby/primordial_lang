# `args()` Builtin Research and Delivery Plan

Status: proposed

Target signature: `args(): []string`

## Executive recommendation

Implement `args()` as a zero-argument builtin that returns a fresh slice containing only the arguments intended for the Primordial program. Do not include the Primordial executable, the `run` subcommand, or the source-file path.

For example:

```text
primordial run proxy.pri -- :8080 http://localhost:3000
```

should expose:

```primordial
args() == []string{":8080", "http://localhost:3000"}
```

The launcher should own argument separation and inject an immutable snapshot into the runtime. The builtin should not read `os.Args` directly. Both the semantic analyser and evaluator should consume one shared builtin definition so they cannot disagree about the builtin's existence, arity, or return type.

## Why this is larger than a normal function

`args()` crosses four boundaries:

```text
host process
    → Primordial launcher
    → builtin/runtime boundary
    → Primordial []string value
```

It also exposes an existing architectural gap. Runtime builtins currently live in [`evaluator/builtins.go`](../../../../evaluator/builtins.go), while [`semantic.NewSymbolTable`](../../../../semantic/symbol_table.go) starts empty. As a result, evaluation can resolve `len`, but normal semantic analysis has no equivalent builtin definition. The current evaluator tests do not reveal this consistently because their helper runs semantic analysis without checking its returned errors.

The `args()` feature should therefore establish a reusable builtin-registration path rather than adding a second evaluator-only global.

## Research findings

### Host argument conventions

Go exposes `os.Args` starting with the host program name. Its `flag.FlagSet.Parse` API deliberately accepts a slice that excludes the command name, demonstrating a useful separation between raw process arguments and application operands. See the official [`os.Args`](https://pkg.go.dev/os#Args) and [`FlagSet.Parse`](https://pkg.go.dev/flag#FlagSet.Parse) documentation.

Python exposes a program-oriented `sys.argv` and separately preserves interpreter arguments in `sys.orig_argv`. This is useful precedent for not leaking launcher implementation details into the user program. See [`sys.argv`](https://docs.python.org/3/library/sys.html#sys.argv) and [`sys.orig_argv`](https://docs.python.org/3/library/sys.html#sys.orig_argv).

Node exposes the executable and script in the first two `process.argv` positions, while also separating runtime-specific options into `process.execArgv`. This again shows that launcher and user arguments are distinct concepts, even though Node chooses to retain both leading entries. See [`process.argv`](https://nodejs.org/api/process.html#processargv).

Primordial should choose the cleaner application-facing contract: `args()[0]` is the first user-supplied argument.

### `--` and argument preservation

POSIX Utility Syntax Guideline 10 defines `--` as the end-of-options delimiter, after which values beginning with `-` are operands. Primordial should accept an optional `--` between the source path and program arguments. See the [POSIX utility syntax guidelines](https://pubs.opengroup.org/onlinepubs/9699919799/basedefs/V1_chap12.html).

The launcher must pass each already-separated host argument through as one Primordial string. It must not join and re-split arguments, reinterpret quotes, or expand wildcards. Shells and host platforms have already performed their argument processing by the time Go supplies `os.Args`.

### Text encoding

Primordial strings are intended to be UTF-8 text, but operating-system arguments are not universally guaranteed to contain valid Unicode. Rust makes this distinction explicit: `std::env::args` fails on invalid Unicode while `args_os` preserves platform-native strings. See [`std::env::args`](https://doc.rust-lang.org/std/env/fn.args.html) and [`std::env::args_os`](https://doc.rust-lang.org/std/env/fn.args_os.html).

Primordial does not currently have an `OsString` or byte-string type. The launcher should therefore validate arguments with Go's [`utf8.ValidString`](https://pkg.go.dev/unicode/utf8#ValidString) and report a launcher error before parsing or evaluating the program. Lossy conversion would be unsafe for filenames and proxy configuration.

## Proposed language contract

```primordial
args(): []string
```

Rules:

1. `args()` accepts no arguments.
2. It returns only user-program arguments.
3. The returned order exactly matches the launcher input.
4. Empty strings and strings containing spaces are preserved.
5. With no supplied arguments, it returns a non-nil empty slice.
6. Each call returns an independent slice so program mutation cannot change the runtime snapshot or later results.
7. The REPL supplies an empty argument snapshot.
8. Invalid UTF-8 is rejected by the launcher.
9. The source path is not included. A future `programPath()` capability can expose it explicitly if needed.
10. `args` remains a normal prelude name rather than a keyword. User code may shadow it in an inner/current scope according to normal lexical rules.

## Proposed architecture

### Shared builtin definition

Introduce a small package that does not depend on `evaluator` or `semantic`. A definition needs enough information for both phases:

```go
type RuntimeContext struct {
    Args []string
}

type Definition struct {
    Name      string
    CheckCall func(argumentTypes []types.Type) (types.Type, error)
    Bind      func(RuntimeContext) *object.Builtin
}
```

`CheckCall` is preferable to a fixed parameter list because existing builtins such as `len` are polymorphic over strings, arrays, and slices. A shared registry also prevents a builtin from working at runtime while being rejected semantically.

### Prelude scopes

Build two parallel preludes from the same definitions:

```text
builtin definitions
    ├── semantic prelude → BuiltinSymbol
    └── runtime prelude  → object.Builtin closures
```

The user program's root scope/environment should be enclosed by the prelude rather than sharing the same map. This preserves normal lexical lookup and allows an explicit user binding to shadow a builtin without mutating the prelude.

### Runtime binding

The `args` runtime function should close over a copied `RuntimeContext.Args`. Every invocation should allocate a new `object.Slice` and populate it with `object.String` values. It should still validate its runtime arity defensively, even though valid programs have already passed semantic analysis.

### Launcher boundary

The launcher should eventually support:

```text
primordial run [runner-options] <source.pri> [--] [program-arguments...]
```

Everything after the source path is passed through. If the first remaining token is `--`, the launcher consumes that token and passes the rest. Keeping this logic outside the evaluator makes the builtin deterministic in tests and reusable by a future compiler, WASI runtime, or embedded host.

## Rejected alternatives

### Read `os.Args` inside the builtin

This is quick but couples language behaviour to the current Go executable layout, leaks `run` and the source path, makes tests depend on the test process, and makes embedding harder.

### Store arguments in a mutable package global

This introduces test interference, prevents safe concurrent evaluator instances, and makes runtime state implicit.

### Add only an evaluator builtin

The semantic analyser would continue rejecting valid calls or returning an unknown function type. Runtime and semantic definitions must be introduced together.

### Include the program path at index zero

This follows C, Go, Rust, and Python traditions, but the value is platform- and launcher-dependent. Primordial can expose program metadata through a separately named capability later.

## Delivery order

| Order | Ticket | Outcome |
|---:|---|---|
| 1 | [ARG-001](001-define-language-contract.md) | Stable language and launcher contract |
| 2 | [ARG-002](002-add-runtime-context.md) | Injectable, immutable host inputs |
| 3 | [ARG-003](003-create-builtin-registry.md) | One source of truth for builtin behaviour |
| 4 | [ARG-004](004-integrate-semantic-prelude.md) | `args()` type-checks as `[]string` |
| 5 | [ARG-005](005-bind-runtime-args.md) | Evaluator returns the runtime slice |
| 6 | [ARG-006](006-add-run-command.md) | Real `.pri` programs receive CLI arguments |
| 7 | [ARG-007](007-add-integration-coverage.md) | End-to-end guarantees and user documentation |

ARG-002 and ARG-003 can be implemented independently after ARG-001. ARG-004 and ARG-005 depend on ARG-003. ARG-006 depends on ARG-002 and ARG-005. ARG-007 closes the feature.

## Definition of done

The feature is complete when this program passes semantic analysis and evaluates using arguments supplied by the launcher:

```primordial
arguments := args();
len(arguments);
```

and:

```text
primordial run example.pri -- one "two words" -three
```

produces a Primordial slice equivalent to:

```primordial
[]string{"one", "two words", "-three"}
```

without exposing launcher tokens or sharing mutable slice storage across calls.
