# PUSH-002 — Add member-expression syntax

## Description

Teach the lexer, AST, and Pratt parser to represent method-call syntax. Keep the
new node generic so later slice methods do not require parser changes.

## Technical description

Add a token for `.` and lex it independently from numeric literals. Introduce an
AST member expression containing the receiver expression and member identifier.
Register `.` as a high-precedence infix/postfix parse function.

Parse:

```primordial
values.push(1).push(2)
```

as nested call expressions whose callees are member expressions. The existing
call parser should remain responsible for parentheses and arguments. Member
access must bind tightly enough to coexist predictably with indexing and calls,
including `values[0].method()` when that receiver type eventually supports a
method.

Do not introduce a push-specific AST node or accept arbitrary expressions after
the dot. Bare member access can parse, but semantic analysis will reject method
values in the first version.

## Acceptance criteria

- The lexer emits a dedicated token for `.`.
- The AST has a reusable member-expression node with receiver and member name.
- `values.push(1)` parses as a call whose function is a member expression.
- `values.push(1).push(2)` parses left-associatively as two chained calls.
- Member access composes with call and index postfix expressions.
- Missing identifiers after `.` produce a focused parser error.
- Existing global function calls and index expressions parse unchanged.
- Lexer, AST stringification, and parser tests cover the new syntax.
