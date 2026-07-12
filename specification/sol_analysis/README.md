# Primordial: language analysis, working specification, and guide

Status: design proposal based on the repository as of 11 July 2026.

This document describes the language Primordial appears to be becoming. It is deliberately split into three levels:

- **Implemented** means the lexer/parser or semantic analyser accepts the feature today.
- **Designed** means the root README specifies the feature, but the implementation does not yet provide it end to end.
- **Proposed** is an extrapolation in this document. It is not a claim about the author's settled intent.

Where the implementation, tests, and README disagree, this document reports the disagreement rather than silently choosing one as fact.

## 1. Executive analysis

Primordial's strongest identity is not simply “Rust + Go + Zig + JavaScript.” It is a small, expression-oriented, statically typed language with:

- Go-like declaration and package ergonomics;
- Rust-like immutable-by-default bindings and explicit fallibility;
- Zig-like visible control flow and a preference for explicit allocation/error policy;
- JavaScript-like first-class functions, closures, object-friendly programming, and low-friction scripting syntax.

The current implementation is a typed Monkey-style interpreter core: integer and boolean literals, arithmetic/comparison, prefix operators, `if` expressions, declarations, function declarations/literals, calls, and multiple return statements. A semantic pass adds scopes, type checking, immutability data, call validation, and return validation. The evaluator currently executes only the expression/control-flow subset; it does not execute declarations, assignments, functions, identifiers, or calls.

The language's most promising design centre is:

> Make important effects visible without making ordinary code ceremonial.

That principle explains immutable-by-default local bindings, expression-valued conditionals, typed functions, explicit `Result`, and concise inference. It should also guide future choices: avoid hidden numeric conversions, hidden exceptions, hidden allocation, and hidden mutation, while keeping the common syntax compact.

## 2. What exists today

### 2.1 Front-end pipeline

The repository contains:

1. a byte-oriented lexer with line/column fields;
2. a Pratt parser producing an AST;
3. a semantic analyser with nested symbol tables;
4. a small tree-walking evaluator and REPL;
5. a type model intended to grow toward compilation;
6. an eventual LLVM compilation goal stated in the README.

The semantic analyser is currently more capable than the evaluator. “Accepted by analysis” must not yet be read as “runnable.”

### 2.2 Implemented surface

| Area | Current behaviour | Important limitation |
|---|---|---|
| Identifiers | ASCII letters/underscore internally, then digits | Leading `_` is lexed as illegal despite helper accepting it; naming style is not enforced |
| Literals | decimal integers, `true`, `false` | String token exists but strings are not lexed or parsed; no float literals |
| Operators | `+ - * /`, `!`, `== != < >`; lexer also recognizes `<= >=` | `<=` and `>=` are not parsed/analyzed/evaluated |
| Declarations | `[pub] [mut\|const] name [: type] := expression` | Runtime evaluation absent; malformed declarations often disappear instead of producing useful parser errors |
| Assignment | AST and semantic rule exist | Parser never constructs assignment statements; evaluator absent |
| Conditionals | `if (condition) { ... } else if (...) { ... } else { ... }` | Semantic analysis requires boolean conditions, evaluator treats any non-false value as truthy |
| Functions | named declarations and anonymous literals, typed parameters, zero or more return types | Calls analyze only identifier callees; evaluator does not execute functions |
| Returns | comma-separated values | Empty `return;` works accidentally via parser flow; call sites reject multiple results |
| Types | `bool`, fixed-width integer/unsigned/float names, `string`, `function`, `void`, internal tuple/named types | Lexer has dedicated tokens for only some type names, although parser treats all type names as identifiers |
| Scopes | nested function scopes | Blocks do not create semantic scopes; closure capture/runtime environment not implemented |

### 2.3 Operator precedence

From tightest to loosest, the parser currently uses:

1. calls: `f(x)`
2. prefix: `!x`, `-x`
3. product: `*`, `/`
4. sum: `+`, `-`
5. ordering: `<`, `>`
6. equality: `==`, `!=`

All binary operators are left-associative.

### 2.4 Known contradictions and gaps

These should be resolved before calling any document a normative v0.1 specification:

- The README names `boolean`; implementation and tests use `bool`.
- The README lists `int` and `uint`; implementation only exposes fixed-width forms.
- The lexer defines `STRING_LITERAL`, but has no quote handling.
- The lexer emits `<=` and `>=`, but the parser has no precedence or infix registration for them.
- The AST and analyser support assignment, but the parser does not.
- The semantic analyser requires a boolean `if` condition; evaluator tests explicitly accept `if (1)`. A statically typed language should choose the semantic rule and remove truthiness.
- Multi-value returns are accepted in declarations and `return`, but a call with multiple results is rejected even though a tuple type exists.
- `pub` and `const` are stored but have no module or compile-time semantics yet.
- Function symbols retain signatures, while general function values collapse to a generic `function` type. Higher-order calls therefore cannot be checked precisely.
- Function return-path analysis marks a block as returning if it contains a return anywhere, not only when every reachable path returns.
- The current baseline `go test ./...` fails in `ast.TestString` because an inferred declaration with a nil `Type` is rendered as if explicitly typed.

## 3. Proposed normative core (v0.1)

This section is a proposed coherent contract. It keeps existing syntax where possible and resolves ambiguities conservatively.

### 3.1 Source files and lexical rules

Primordial source files use the `.pri` extension and UTF-8 encoding.

Whitespace separates tokens and is otherwise insignificant. Newlines do not terminate statements. Semicolons terminate simple statements in v0.1; a later formatter may allow safe semicolon insertion, but the grammar should not depend on that initially.

Line comments use `//`. Block comments should be deferred until nesting and documentation comments are designed.

Identifiers should follow this lexical grammar:

```text
identifier = letter { letter | decimal_digit | "_" } ;
letter     = "A"…"Z" | "a"…"z" | "_" ;
```

Style guidance, not a parse error:

- local names and functions: `camelCase`;
- types: `PascalCase`;
- constants: ordinary `camelCase` rather than a second naming language;
- a leading underscore suppresses an unused-binding diagnostic.

Keywords proposed for the first useful compiled language:

```text
fn pub mut const if else return true false
type struct enum match import as try defer
```

Reserve a word only when its grammar is introduced. Premature reservation makes experimentation harder.

### 3.2 Values and types

Recommended primitive types:

```text
bool
i8 i16 i32 i64 isize
u8 u16 u32 u64 usize
f32 f64
string
void
```

For compatibility with the current implementation, `int8`/`uint8` forms may remain temporarily, but short `i32`/`u32` spellings fit the Rust/Zig adjacency and reduce noise. Pick one family before v0.1; do not permanently support two aliases for every integer type.

Integer literals should begin as **untyped integer constants** and acquire a concrete type from context. With no constraining context, default to `i64` initially because that matches the current analyser and evaluator. This preserves `x: i32 := 1` without permitting arbitrary integer-to-integer assignment.

Important rule: `i32` is not implicitly assignable to `u64`, and `i64` is not implicitly assignable to `i8`. The current `IsAssignable` accepts every integer pair, which can silently truncate or change sign. Require an explicit checked conversion:

```pri
small := i8.from(large)?;
```

The exact conversion syntax can wait; the no-hidden-narrowing rule should not.

Proposed compound types:

```text
(T, U)             tuple
[N]T               fixed array
[]T                slice/view
map[K]V            hash map, after generics exist
fn(T, U): R        function
T?                 optional
Result[T, E]       result
```

`string` should be immutable UTF-8 text. Indexing should either return a byte explicitly or be omitted; pretending constant-time indexing yields a Unicode character repeats a common JavaScript usability trap.

### 3.3 Bindings, constants, and assignment

Grammar:

```text
declaration = [ "pub" ] [ "mut" | "const" ] identifier
              [ ":" type ] ":=" expression ";" ;
assignment  = place "=" expression ";" ;
```

Examples:

```pri
port := 8080;                  // immutable runtime binding, inferred
mut attempts: i32 := 0;       // mutable runtime binding
const maxRetries: i32 := 3;   // compile-time value
pub const version := "0.1";   // exported compile-time value
```

Rules:

1. Every binding is initialized.
2. A plain binding is immutable.
3. `mut` permits reassignment of that binding or mutation through that binding when the value's API permits it.
4. `const` requires compile-time evaluation and cannot be combined with `mut`.
5. `pub` controls package visibility; it does not change mutability.
6. `:=` always declares in the current scope. It never conditionally reuses an outer binding.
7. `=` only assigns to an existing mutable place.
8. Shadowing should be allowed only with an explicit `shadow` construct or forbidden in v0.1. Silent shadowing is concise but too easy to confuse with assignment.

Mutability should be lexical, not “local package” authority. A function may mutate a package-level variable only if that variable is `mut` and visible, but unrestricted shared package mutation will eventually conflict with concurrency and optimisation. Prefer passing mutable state explicitly.

### 3.4 Blocks and scope

Every `{ ... }` block creates a lexical scope. Bindings live from their declaration to the end of that block. Inner scopes may read outer bindings. Capturing an immutable outer binding in a function literal is allowed. Capturing mutable state should initially use shared GC semantics; later concurrency rules may restrict it.

A block's value is its final expression without a terminating semicolon:

```pri
answer := {
    base := 40;
    base + 2
};
```

A final semicolon discards the expression and makes the block `void`. This Rust-like distinction gives expression-oriented syntax a precise rule and removes the current ambiguity where semicolon presence is discarded by the AST.

### 3.5 Expressions and operators

Proposed precedence, tightest first:

| Level | Operators |
|---|---|
| postfix | call `()`, field `.`, index `[]`, propagation `?` |
| prefix | `!`, unary `-`, unary `+` (optional) |
| multiplicative | `* / %` |
| additive | `+ -` |
| ordering | `< <= > >=` |
| equality | `== !=` |
| logical and | `&&` |
| logical or | `||` |

Logical operators short-circuit and accept only `bool`. Primordial should have no general truthiness: `0`, empty strings, empty collections, and optional values are not booleans.

Equality is available only for types whose equality semantics are defined. Function values are not comparable. Floating-point equality remains exact and should trigger a lint in suspicious code rather than having surprising language semantics.

### 3.6 Conditional expressions

Syntax:

```pri
label := if (score >= 90) {
    "excellent"
} else if (score >= 70) {
    "good"
} else {
    "developing"
};
```

Rules:

1. The condition has type `bool`.
2. In statement position, `else` is optional and branch values are discarded.
3. In value position, all reachable non-diverging branches yield compatible types.
4. A value-position conditional is exhaustive and therefore requires a final `else`.
5. A branch that returns, breaks, or otherwise diverges does not need to yield the value type.

Parentheses around the condition are currently required. They could later become optional, but retaining them gives familiar JavaScript syntax and makes the parser unambiguous while the language is young.

### 3.7 Functions and closures

Canonical grammar:

```text
function_decl = [ "pub" ] "fn" identifier signature block ;
function_lit  = "fn" signature block ;
signature     = "(" [ parameters ] ")" [ ":" return_type ] ;
parameters    = parameter { "," parameter } ;
parameter     = identifier type ;
```

Examples:

```pri
fn add(x i32, y i32): i32 {
    x + y
}

double := fn(value i32): i32 { value * 2 };

fn log(message string) {
    print(message);
}
```

The type of `add` is `fn(i32, i32): i32`, not generic `function`. Function signatures must survive inference so higher-order code can be type checked:

```pri
fn map(values []i32, transform fn(i32): i32): []i32 { ... }
```

For v0.1, choose one return model: a function returns exactly one type, and multiple values are represented by a tuple. Thus:

```pri
fn bounds(values []i32): (i32, i32) {
    (min, max)
}

(low, high) := bounds(values);
```

This reconciles Go's pleasant destructuring with Rust/Zig's simpler “one expression has one type” model. It also prevents a separate multi-return mechanism from infecting call and generic typing.

`return expression;` exits early. The final expression is the normal result. A `void` function may use `return;`.

Closures capture lexical bindings. GC makes immutable captures straightforward. Mutation of a captured `mut` binding is allowed in the single-threaded v0.1 runtime but should be represented as a heap cell so compiled behaviour matches interpreted behaviour.

### 3.8 Structs, enums, and pattern matching (proposed)

These are the most valuable next data-model features because the stated game/server goals need named data and state machines.

```pri
struct User {
    id u64,
    name string,
    email string?,
}

enum Direction {
    north,
    east,
    south,
    west,
}

enum LoadState[T, E] {
    idle,
    loading,
    ready(T),
    failed(E),
}
```

Prefer algebraic enums over JavaScript-style tagged object conventions. They make invalid states harder to represent and provide the natural foundation for `Option` and `Result`.

```pri
message := match state {
    .idle => "waiting",
    .loading => "loading",
    .ready(value) => value,
    .failed(err) => err.message,
};
```

`match` must be exhaustive in value position. A wildcard arm `_` is allowed but should produce a lint when it hides newly added enum variants across package boundaries.

Methods can be ordinary functions in a type namespace rather than requiring inheritance:

```pri
fn User.displayName(self): string { self.name }
```

No classes, inheritance, implicit constructors, or prototype mutation are recommended. Composition plus enums, structs, and functions covers the intended domains with less implementation weight.

### 3.9 Optionals

Use `T?` for a value that may be absent, backed by an enum-like representation rather than `null`:

```pri
middleName: string? := none;

name := match middleName {
    some(value) => value,
    none => "",
};
```

The eventual convenience syntax should be small:

- `value?` propagates absence only when the enclosing function returns an optional;
- `value ?? fallback` unwraps with a fallback;
- `if value |present| { ... }` is a possible Zig-like payload capture, but `match` may already be sufficient.

Do not add both JavaScript-style optional chaining and several separate optional-binding syntaxes immediately. Start with `match`, then add only the sugar proven useful by real programs.

### 3.10 Errors and `Result`

Fallible functions return `Result[T, E]`. Errors are values and do not unwind through exceptions.

```pri
enum UserError {
    notFound,
    unavailable(string),
}

fn getUser(id u64): Result[User, UserError] {
    ...
}
```

Use postfix `?` as the primitive propagation operator:

```pri
fn displayName(id u64): Result[string, UserError] {
    user := getUser(id)?;
    ok(user.name)
}
```

The README proposes prefix `try`. Either can work, but postfix `?` scales better through expression chains and aligns with Rust while preserving explicit control flow. If the author strongly prefers `try`, define it as syntax sugar for the same operation and avoid having both permanently.

Handle errors exhaustively with `match`:

```pri
match getUser(id) {
    ok(user) => render(user),
    err(.notFound) => render404(),
    err(problem) => log(problem),
};
```

Not every value-returning function should automatically be wrapped in `Result`. Fallibility belongs in the declared return type. Automatically wrapping all functions creates ceremony and obscures what can fail.

The proposed `retry` should begin as a standard-library function, not privileged syntax:

```pri
user := retry(policy, fn(): Result[User, UserError] {
    api.getUser(id)
})?;
```

This is testable, composable, and does not require the compiler to know about clocks. A later trailing-closure syntax could recover the README's ergonomic form. Retry policies should include maximum attempts, backoff, jitter, and a predicate identifying retryable errors. The default must not retry arbitrary errors.

### 3.11 Memory and resource management

The stated GC choice is sensible for the project's scale and JavaScript-like closure ergonomics. Do not add a partial borrow checker: a less-complete ownership system can be harder to reason about than a good GC.

Recommended model:

- tracing GC for ordinary heap objects and closures;
- value semantics for small primitives and structs where practical;
- explicit references/slices for shared views;
- deterministic cleanup with `defer` for files, sockets, locks, and foreign resources;
- no finalizers as the primary resource-management mechanism.

```pri
file := fs.open(path)?;
defer file.close();
```

This takes Go/Zig's practical cleanup visibility without Rust lifetime syntax. If performance work later demands arenas, add an allocator/arena library whose use is explicit at subsystem boundaries rather than changing every type.

### 3.12 Packages and imports

Folder-based packages suit the Go adjacency, but the README's ancestor restriction is likely too limiting for real applications. A nested package commonly needs shared types, configuration, logging, or its parent API. The rule also makes code movement change which imports are legal.

Recommended rules:

1. a package corresponds to a directory;
2. any package in the same module may import any other package;
3. the dependency graph must be acyclic;
4. only `pub` names cross package boundaries;
5. imports are explicit and may be aliased;
6. unused imports are compile errors;
7. package initialization is either absent in v0.1 or has a deterministic dependency order.

```pri
import ecom/auth/middleware as auth;
import ecom/model;

user: model.User := auth.requireUser(request)?;
```

Avoid treating an imported package as a runtime map. It should be a compile-time namespace. This enables static lookup, dead-code elimination, and clear LLVM symbols.

### 3.13 Visibility

Use explicit `pub`; do not infer export from capitalization. This is already the better part of the current design because naming style and API surface remain separate decisions.

Recommended visibility levels initially:

- no modifier: package-private;
- `pub`: module/public API.

Do not add Rust's full `pub(crate)` family until separate modules and packages create a concrete need.

### 3.14 Generics and interfaces (later proposal)

Generics are valuable for `Result`, collections, iterators, and reusable game/server components, but should follow named types and functions rather than precede them.

```pri
fn first[T](items []T): T? { ... }
```

Start with monomorphized parametric generics. It maps cleanly to LLVM, provides predictable performance, and avoids a dynamic “anything” escape hatch.

For behavioural abstraction, prefer small structural interfaces/traits satisfied implicitly, Go-style:

```pri
interface Writer {
    write(bytes []u8): Result[usize, IoError];
}
```

However, make dynamic dispatch explicit in the type, for example `dyn Writer`, while generic `T: Writer` uses static dispatch. This borrows Rust's useful distinction without borrowing lifetimes.

Avoid a universal `any` in the core language. If interop requires one, make downcasts explicit and fallible.

### 3.15 Concurrency and async (defer)

Do not commit to goroutines, actors, or JavaScript promises until the memory and error models are stable. For the server goal, structured concurrency is the most coherent eventual direction:

- tasks have an owning scope;
- cancellation flows from parent to child;
- task failure is a `Result`, not an unhandled exception;
- mutable shared state requires an explicit synchronization type;
- detached tasks are rare and explicit.

An `async fn`/`await` surface is familiar, but its runtime should not be designed until synchronous I/O and cleanup work. Go-like channels could later be a library abstraction rather than syntax.

## 4. How to write Primordial

This section shows the proposed language, not the exact current interpreter.

### 4.1 A small program

```pri
import std/io;

fn classify(value i32): string {
    if (value < 0) {
        "negative"
    } else if (value == 0) {
        "zero"
    } else {
        "positive"
    }
}

fn main(): Result[void, io.Error] {
    const target: i32 := 42;
    label := classify(target);
    io.println(label)?;
    ok(void)
}
```

### 4.2 State and mutation

```pri
fn countMatches(values []i32, target i32): usize {
    mut count: usize := 0;

    for value in values {
        if (value == target) {
            count = count + 1;
        }
    }

    count
}
```

`for` and slices are proposed, but this example illustrates the intended rule: mutation is local and opt-in.

### 4.3 Data modelling

```pri
struct Character {
    name string,
    health i32,
}

enum Action {
    wait,
    attack(usize),
    useItem(usize),
}

fn isAlive(character Character): bool {
    character.health > 0
}
```

This is enough to model an early RPG engine without inheritance or an object system.

### 4.4 Recoverable failure

```pri
fn loadCharacter(path string): Result[Character, LoadError] {
    source := fs.readString(path)?;
    dto := json.parse[CharacterDto](source)?;
    validate(dto)
}
```

Each `?` is a visible early return. Callers can propagate again or exhaustively handle the error.

## 5. Adjacency: what to borrow and what not to borrow

### 5.1 Rust

Borrow:

- immutable bindings by default and explicit `mut`;
- expression-valued blocks and conditionals;
- algebraic enums and exhaustive matching;
- `Result`, `Option`, and propagation with `?`;
- precise function types and monomorphized generics.

Do not borrow initially:

- explicit lifetimes and a borrow checker;
- a large trait-resolution system;
- macro complexity before the base grammar is stable.

The official Rust Book documents immutable-by-default variables and `mut`, and its recoverable-error chapter defines `Result` and `?`: [variables and mutability](https://doc.rust-lang.org/book/ch03-01-variables-and-mutability.html), [recoverable errors with `Result`](https://doc.rust-lang.org/book/ch09-02-recoverable-errors-with-result.html).

### 5.2 Go

Borrow:

- directory packages, explicit imports, and cycle rejection;
- small interfaces satisfied without declaration;
- a formatter as part of the standard toolchain;
- simple build/test workflows;
- `defer` and visible control flow.

Adapt rather than copy:

- retain `:=`, but make it declaration-only and avoid Go's mixed “some new, some existing” rule;
- represent multiple values as tuples rather than a special multi-return channel;
- use `Result` propagation rather than repetitive `if err != nil`;
- use conventional generic brackets rather than making syntax novelty a goal.

The Go specification is the primary reference for short declarations, function types/results, packages, interfaces, and initialization: [Go Language Specification](https://go.dev/ref/spec).

### 5.3 Zig

Borrow:

- errors as part of types and explicit propagation;
- optionals distinct from pointers and ordinary values;
- `defer` for lexical cleanup;
- explicit allocators/arenas when a subsystem needs allocation control;
- compile-time evaluation as a capability of ordinary code rather than a separate macro language.

Do not borrow initially:

- pervasive allocator parameters in everyday application code;
- unrestricted compile-time reflection before type checking is mature;
- syntax whose cleverness outweighs familiarity for the intended solo-project workflow.

The Zig language reference documents error unions, optionals, `try`, `defer`, and `comptime`: [Zig Language Reference](https://ziglang.org/documentation/master/).

### 5.4 JavaScript

Borrow:

- first-class functions and lexical closures;
- concise object/collection literals;
- approachable expression syntax;
- productive REPL and scripting experience;
- a module experience that feels direct to use.

Reject:

- truthiness and coercive equality;
- prototype mutation as the primary data model;
- `null`/`undefined` overlap;
- exceptions as the ordinary recoverable-error path;
- implicit numeric/string conversion.

JavaScript closures demonstrate why lexical functions are so enabling, while Primordial can retain them with static capture types: [MDN JavaScript closures](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Closures).

## 6. Recommended implementation sequence

The shortest route from the current repository to a coherent usable language is:

1. **Stabilize the front end.** Preserve semicolons in the AST, report all parser errors, implement comments/strings/floats, add `<= >=`, parse assignment, and reconcile type spellings.
2. **Make the interpreter honest.** Add environments, declarations, identifiers, assignment, functions, closures, calls, and evaluator/semantic agreement on boolean conditions.
3. **Unify return values.** Implement tuple expressions/types/destructuring and make functions return one type.
4. **Add named data.** Struct declarations, literals, field access, enums, and `match` unlock meaningful programs.
5. **Add `Option` and `Result`.** Implement them using the enum/type machinery, then add propagation sugar.
6. **Build packages.** Module discovery, import graph, `pub`, name resolution, and cycle diagnostics.
7. **Choose runtime memory representation.** Establish strings, arrays, closures, GC roots, and FFI boundaries before LLVM lowering.
8. **Introduce LLVM incrementally.** Lower a typed, desugared IR rather than lowering the parser AST directly. Keep interpreter conformance tests as the semantic oracle.
9. **Only then add generics and concurrency.** Both multiply the state space of every earlier feature.

The type checker deserves immediate hardening alongside step 1:

- preserve concrete function signatures;
- use untyped constants rather than broad integer assignability;
- create scopes for all blocks;
- model control-flow divergence accurately;
- prevent use-before-declaration and invalid `return` at top level;
- make every analyser error recover without panicking or generating cascades where possible.

## 7. Proposed conformance principles

Each language feature should have tests at four boundaries:

1. lexical tokens and source positions;
2. parsed AST and diagnostics;
3. resolved/type-checked representation and diagnostics;
4. evaluator and compiled result equivalence.

Good conformance cases include success, syntax failure, type failure, scope failure, and boundary values. An accepted program must mean the same thing in the interpreter and LLVM backend. If the interpreter intentionally supports a debug-only extension, it should use a separate mode rather than silently diverge.

Diagnostics should include file, line, column, a short message, the relevant source span, and one useful note. Invalid input should never be converted into a missing AST node without a recorded error.

## 8. Design decisions to settle explicitly

These are the few decisions with enough downstream impact to deserve an explicit author choice:

1. `bool` versus `boolean`.
2. `i32`/`u32` versus `int32`/`uint32`, and the meaning of unsized `int`/`uint` if retained.
3. mandatory semicolons versus Rust-like final-expression semicolon significance.
4. prefix `try` versus postfix `?` for result propagation.
5. whether `Result[T, E]` is ordinary generic enum syntax or a compiler-special type.
6. tuple-return model versus distinct multiple return values.
7. whether implicit integer widening exists; this analysis recommends no implicit conversion except contextual typing of literals.
8. GC-only v0.1 versus user-visible allocators/arenas at selected APIs.
9. unrestricted same-module imports versus the current ancestor restriction.
10. whether shadowing is forbidden, explicit, or freely allowed.

Everything else can evolve behind those choices with relatively little syntactic churn.

## 9. A concise language manifesto

Primordial should feel fast to start, difficult to misuse accidentally, and straightforward to finish projects in.

- Values are typed; conversions are visible.
- Bindings are immutable unless marked `mut`.
- Absence and failure are types, not ambient runtime surprises.
- Blocks and conditionals produce values.
- Functions and closures are ordinary values with precise types.
- Data is composed from structs, enums, tuples, and collections.
- Resources have lexical cleanup; ordinary memory uses GC.
- Packages are static namespaces with explicit public APIs.
- The compiler favours clear diagnostics and one unsurprising way to express common work.
- Advanced power arrives through composable libraries and types before new syntax.

That combination is adjacent to Rust, Go, Zig, and JavaScript without becoming a collage of them. It gives Primordial a defensible identity of its own: **explicit where correctness depends on it, concise everywhere else**.
