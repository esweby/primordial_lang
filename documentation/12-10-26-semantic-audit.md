# Semantic Analyzer Audit

With the resolver (#3) and inference (#4) merged, much of the semantic analyzer now duplicates work done in earlier phases. This audit sorts every check and every test in `semantic/` into one of four groups, so the analyzer can be slimmed down to the rule checks that are left.

- **Resolver**: names and scopes, already handled by `resolver/`
- **Inference**: types and all integer rules, already handled by `inference/`
- **Analyzer**: a non-integer rule that stays in `semantic/`
- **Delete**: dead code or internal guards that no longer apply

## Decisions

These came out of the audit and were settled before anything was removed.

1. **Public fields are externally settable.** The language specification already decides this: "Declaring a property as publicly accessible will also make that property publicly settable" (Structs → Properties, rule 2.1). `TestAssignmentAnalysis` is correct, and `TestRejectStructMemberMisuse/external_field_assignment`, which expected assigning a `pub` field from outside to fail, is wrong and is deleted. That contradiction is why the test had been failing.
2. **The `function` annotation stays.** It only loses information when a function passes through a slot typed `function`, such as a parameter, return type or field. A declaration like `add: function := fn(...)` still gets its full signature from the literal. The longer-term improvement is to add signature types such as `fn(int64): int64` as annotations alongside `function`, so callbacks keep their argument and return types. That is future parser work, not part of this slim-down.
3. **An empty program is not an error.** The analyzer's "analyzer invoked on empty program" check is deleted, matching the resolver and inference.

## Checks

### Already handled by the resolver

| Check | Where in semantic |
| --- | --- |
| Variable already declared in this scope | statements.go:13, collections.go:96 |
| Type already declared in this scope | structs.go:250 |
| Name appears twice in a tuple declaration | collections.go:90 |
| Undefined identifier, variable or function | expressions.go:62, 250; statements.go:104; collections.go:136 |
| Unknown struct in a struct literal, or the name is not a struct | structs.go:172, 177 |
| Unknown type for a field, parameter or return | structs.go:48, 111, 120 |
| `break` or `continue` outside a loop, or an unknown label | loops.go:189, 196, 224 |
| Range loop needs 1 or 2 variables | loops.go:160, 181 |
| Hoisting structs so they can be used before they are declared | `predeclareStruct`, structs.go:245 |
| Scope tracking: the symbol table, enter/exit scope, the loop depth and label stacks | symbol_table.go, semantic_analyzer.go |

### Already handled by inference

| Check | Where in semantic |
| --- | --- |
| Integer literal does not fit its type | integers.go:66 |
| Untyped integer used as a non-integer type | integers.go:55 |
| Mismatched integer types in operators | integers.go:149; expressions.go:121, 142, 157 |
| Division by a constant zero | expressions.go:126 |
| Negating an unsigned integer | expressions.go:202 |
| Integer conversions: argument count, value fits, convertible type | expressions.go:300, 306, 310 |
| Integer `if` branches and `break` values taking a shared type | expressions.go:386, 392; loops.go:47, 211; integers.go:74–101 |
| Defaulting untyped integers (now to `int`, previously `int64`) | semantic_analyzer.go:63; `defaultInteger` |
| Integer indexes into slices, arrays and maps | `analyzeIndexExpression`, expressions.go:410 |
| The integer half of every "can this value go here" check: declarations, assignments, returns, arguments, elements, map pairs, struct fields and defaults, tuple assignment | `requireAssignable`, integers.go:103 |
| Inferring types of declarations, struct fields from defaults, `len()`, range variables, member and call results | throughout |
| Linking named types to their struct | `resolveType`, structs.go:213 |

Inference also fixes one bug here. The analyzer only accepts a bare literal as a slice index, so `[]int32{0,1,2}[1+1]` is rejected. Inference accepts it, as Go would.

### Stays in the analyzer

These rules are not about names or integers. With every expression now typed, each becomes a check against the types inference recorded instead of working them out again.

| Area | Checks |
| --- | --- |
| Assignability of non-integer types | declaration, assignment, return values, arguments, collection elements, map pairs, struct literal fields, struct defaults, tuple assignment |
| Operators | non-integer mismatches (`"a" + true`), `-` on a non-number, `!` on a non-bool |
| Conditions | `if`, `while` and the condition of a constructed `for` loop must be `bool` |
| `if` as a value | needs an `else`, both branches return, a branch must end in an expression, non-integer branch mismatches |
| Loops as values | needs a `break`, mixes `break` with and without a value, non-integer `break` value mismatches, range over a non-collection |
| Calls | calling a non-function, wrong number of arguments, `len()` argument count and kind |
| Returns | wrong number of values, values returned from a void function, missing returns on some paths |
| Mutability | assigning to an immutable variable or to something that isn't a variable, calling a mutating method on an immutable variable |
| Members | no members, unknown member, private field, unknown type function, method used without a call, property called as a method, wrong argument count, assigning to a method |
| Types as values | `value := Person;` |
| Struct declarations | duplicate field, type function, method or member; field with neither a type nor a default |
| Struct literals | field given twice, unknown field, missing required field |
| Tuples | value is not a tuple, wrong number of names |
| Arrays and maps | length must be above zero, too many elements, literal has no element type, map literal type is not a map |
| Indexing | indexing a type that can't be indexed |

### Delete

| Code | Reason |
| --- | --- |
| `checkIntegerIndex` (expressions.go:440) | Never called |
| `isInvalidInfixType` (expressions.go:476) | Never called |
| `integers.go` as a whole | Integer rules now live in inference and `types/integers.go` |
| `symbol_table.go`, `symbols.go`, `Symbols()` | Replaced by `ident.Symbol` and the types on the AST |
| "callee is not an identifier (unsupported)" (expressions.go:243) | Inference types any callable expression, so calls like `getAdder()(1)` no longer need special handling |
| Nil, unexpected-expression and unexpected-statement guards (expressions.go:45, 99; semantic_analyzer.go:115) | Keep only as internal panics if wanted; they are not language rules |
| "unexpected else structure", `analyzeFunctionLiteral` non-function guard | Internal guards the parser already prevents |
| "analyzer invoked on empty program" (semantic_analyzer.go:49) | An empty program is valid (decision 3) |

## Tests

Every semantic test builds its own symbol table, and tests that check a type read it back through `analyzer.Symbols()` (14 calls across four test files). Neither will exist after the slim-down. Kept tests need a new helper that parses, resolves, infers and then analyzes, and reads types off the AST.

Several tests also count errors (`expectedErrors: 3`). Those counts change as checks move between phases, so kept tests should match on messages instead.

| Test | Group | Notes |
| --- | --- | --- |
| `TestAnalyzeInfixExpression` | Inference | `true + 2`, `1 + true`, `1 + fn() {}` are untyped integers used as non-integers; inference covers `"a" + 1` already, so add these cases there |
| `TestAnalyzePrefixExpression` | Analyzer | `!5` and `-true` are non-integer operator rules |
| `TestAnalyzeCallExpression` | Analyzer | `add(true)` is a non-integer argument mismatch |
| `TestAnalyzeIfExpression` | Resolver + Analyzer | the undefined `x`/`y` case is covered by the resolver; its count of 3 errors will change |
| `TestAnalyzeDeclareStatement` | Split | redeclaration → resolver; `{ 1; } else { true; }` → inference; `brian: int32 := true`, `if (2)`, top-level `return` → analyzer |
| `TestAssignmentAnalysis` | Analyzer | valid programs, including assigning a `pub` field from outside (decision 1); `map[string]int` is now valid because `int` exists |
| `TestAnalyzeFunctionLiteral` | Analyzer | `add: int32 := fn() {}` |
| `TestAnalyzeFunctionStatement` | Analyzer | `return x + true` is counted as 2 errors; becomes one analyzer error |
| `TestAnalyzeLoops` | Analyzer | valid programs; keep as a smoke test of the full pipeline |
| `TestAnalyzeBasicMaps`, `TestAnalyzeMapTypeMismatch` | Analyzer | string key into an `int32` map is non-integer |
| `TestIntegerLiteralBoundaries` | Inference | duplicate of inference's overflow tests; move the full boundary table there |
| `TestIntegerLiteralUsesContextSymmetrically` | Inference | covered by `y + 1` / `1 + y` cases |
| `TestContextFreeIntegerDefaultsToInt64` | Inference | now defaults to `int`; covered by `x := 0` |
| `TestRejectMixedConcreteIntegerTypes`, `TestRejectUnsignedNegation` | Inference | covered |
| `TestIntegerConversions` | Inference | covered except `int64(x)` from `int32`; add that case |
| `TestIntegerContextsAcrossLanguageFeatures` | Inference | covered by struct, argument and slice cases |
| `TestAnalyzePartiallyInitializedArray`, `TestAnalyzeEmptyArrayLiteral`, `TestAnalyzeSliceLiteral` | Inference | type checks only (`[5]int32`, `[]int32`); move |
| `TestRejectTooManyArrayElements`, `TestRejectInvalidArrayElementType`, `TestRejectMismatchedDeclaredArrayType` | Analyzer | non-integer rules |
| `TestAnalyzeTupleDeclaration`, `TestAnalyzeTupleAssignment` | Analyzer | valid programs |
| `TestRejectImmutableTupleAssignment` | Analyzer | mutability |
| `TestAnalyzeMemberPropertyTypes`, `TestAnalyzeCollectionMemberMethods` | Inference | `.length` expectations change from `int64` to `int` in stage 4 |
| `TestRejectInvalidMemberUsage` | Analyzer | `1.length` message changes from `int64` to `int` |
| `TestAnalyzeBasicStructs`, `TestAnalyzeStructTypeFunctionsAndInstanceMethods`, `TestAnalyzeNestedStructMethodMutation` | Inference + Analyzer | types move to inference; keep the programs as valid analyzer cases |
| `TestRejectInvalidStructLiterals` | Analyzer | except "wrong type" (`Person{name: 1}`), which is now an inference error |
| `TestInferStructFieldTypesFromDefaults` | Inference | `score = 42` becomes `int` |
| `TestResolveNamedStructTypesInsideCollections` | Inference | covered by named type tests in the resolver and inference |
| `TestRejectInvalidStructFieldDeclarations` | Split | unknown type → resolver (message is now "unknown type: Missing"); `string = 1` → inference; missing type/default and duplicate field → analyzer |
| `TestStructTypeCannotBeUsedAsValue` | Analyzer | |
| `TestAnalyzeStructLiteralShorthand`, `TestStructMethodMayAssignItsField` | Analyzer | valid programs |
| `TestRejectStructMemberMisuse` | Analyzer | delete the `external field assignment` case (decision 1); the other two cases stay |
| `TestRejectPrivateStructFieldAccess` | Analyzer | |

## Order of work

1. Delete `external_field_assignment` (decision 1).
2. Move the integer and type tests listed under Inference into `inference/`, adding the missing cases noted.
3. Switch `.length`, `Slice.IndexType`, and the evaluator's `len()` and range values to `int`, updating the tests that expect `int64`.
4. Wire resolver → inference → analyzer into `main.go` and the REPL. The REPL needs the resolver to keep its scope between submissions, replacing `Symbols()`.
5. Rewrite the analyzer to read types from the AST, delete the checks and code listed above, and port the kept tests to the new helper with message-based assertions.
6. Remove the semantic corpus tests from `resolver/` and `inference/` once semantic depends on them, as the comparison will no longer make sense.
