# ARG-006 — Add file execution and argument forwarding

## Description

Add the launcher path that makes `args()` useful to real Primordial CLI programs.

## Technical description

Extend the executable with:

```text
primordial run [runner-options] <source.pri> [--] [program-arguments...]
```

The launcher should read the complete source file, then lex, parse, semantically analyse, and evaluate it using matching semantic and runtime preludes. Extract all tokens after the source path as program arguments. If the first such token is `--`, consume it.

Validate each program argument with `utf8.ValidString` before constructing the runtime context. Report launcher, file, parser, semantic, and runtime failures distinctly and return non-zero process status for each failure category.

Keep the REPL as the default command temporarily if desired, but ensure its runtime context supplies an empty argument list.

## Acceptance criteria

- `primordial run example.pri` executes a complete source file.
- Arguments after the source path reach `args()` unchanged and in order.
- A leading `--` after the source path is consumed rather than exposed.
- Arguments beginning with `-` after the delimiter are preserved.
- Quoted values received from the host remain one argument and are not re-split.
- Launcher and source-file tokens are not visible through `args()`.
- Invalid UTF-8 is rejected before evaluation with a clear diagnostic.
- File, parser, semantic, and runtime failures result in non-zero exit status.
- The REPL continues to start with an empty argument context.
