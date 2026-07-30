# ARG-002 — Add an injectable runtime context

## Description

Introduce an explicit home for host-provided process data so evaluator builtins do not read mutable package globals such as `os.Args`.

## Technical description

Add a small runtime context containing an argument snapshot:

```go
type RuntimeContext struct {
    Args []string
}
```

Provide a constructor that defensively copies the supplied slice. The context should be passed when runtime builtins are bound, not threaded through every recursive `Eval` call. This keeps the current evaluator API small while allowing future capabilities such as standard streams, environment variables, clocks, and filesystem access to use the same injection boundary.

The context must not import parser, semantic, or evaluator packages.

## Acceptance criteria

- A runtime context type exists and contains program arguments.
- Its constructor defensively copies the caller's input slice.
- Tests prove that mutating the original Go slice does not change the context.
- Tests prove that separately created contexts do not share argument storage.
- The context does not read `os.Args`.
- The context introduces no import cycle.
- The design leaves a documented extension point for future host capabilities.
