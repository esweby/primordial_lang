# Primordial Language

Primordial is a hobby programming language built to test the limits of my brain as a developer. The long-term goal is to produce a language that compiles through LLVM, feels good to write, and gives me an excuse to learn far more about compilers than is probably sensible.

It borrows heavily from Go, Rust, and JavaScript: Go's no-nonsense simplicity, Rust's safety and purposeful decision-making, and JavaScript's ability to let you sit down and make something fun. There are thoughts from Zig in here too. Honestly, I could probably just learn Zig, but I don't want to. I want to make something unique and my own.

Primordial source files use the `.pri` extension.

## What I want from it

- Strong types and sensible safety without lifetimes taking over my entire day.
- Simple code that keeps the problem more interesting than the language.
- Useful tools for the developer without becoming loosey-goosey.
- Errors and important decisions handled where they happen.
- Speed. Obviously.

The language is immutable by default, expression-oriented, and intended to make important behaviour visible without making ordinary code ceremonial.

```pri
fn add(x int32, y int32): int32 {
    return x + y;
}

answer := add(20, 22);
```

## Where it is now

Primordial is being written in Go, using the Monkey interpreter as a loose starting point. The project currently has a lexer, Pratt parser, semantic analyser, tree-walking evaluator, and a small REPL. It is still early and under active construction: rough edges, unfinished features, and changing syntax are all part of the deal.

The eventual destination is an LLVM-backed compiler. Along the way I would like to use Primordial to make a tiny game engine—think early *Final Fantasy* or *Final Fantasy Tactics*—a web server, and whatever other small projects seem entertaining enough to expose the next bad language decision.

## Having a look

Run the REPL:

```sh
go run .
```

Run the test suite:

```sh
go test ./...
```

The original [design notes](specification/Design%20Notes.md) contain the longer brain-dump behind the language. There is also a more structured [language specification](specification/Language%20Specification.md). Both are working documents, not sacred texts.

This is an ambitious solo hobby project. Good enough is good enough—until it becomes interesting to make it better.
