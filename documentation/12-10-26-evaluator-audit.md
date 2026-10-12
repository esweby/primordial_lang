# Evaluator Audit

With the resolver, inference and analyzer in place, a program only reaches the evaluator after every name, type and language rule has been checked. This audit sorts the evaluator's runtime checks against that, with two decisions already made:

1. **The evaluator trusts the checked program.** It behaves like a compiler backend: it does not re-check anything the earlier phases guarantee.
2. **Runtime integer coercion is removed.** Every integer has its type fixed at compile time, so fitting an untyped value to its destination at runtime goes against the language's integer rules. Function signature types (`fn(int64): int64`) were added first so calls through function values are typed at compile time.

## Decisions needed

1. **Calling a `function`-typed value.** This is the one place an untyped integer can still reach runtime, because a value annotated `function` has no signature to type its arguments from. Once coercion is removed, nothing would fit `addFive(5)`'s `5` to the real parameter type. Options:
   - **Make it a compile error (recommended).** Now that signature types exist, calling a `function` value would require giving it a signature, e.g. `addFive: fn(int64): int64`. `function` stays as a type for storing and passing functions, but not calling them. Every call is then checked, which is what lets the evaluator trust the program.
   - **Default the arguments to `int`.** This is consistent with the default rule, but the call stays unchecked, so the evaluator would need a runtime argument check just for this case.
2. **Integer overflow at runtime traps.** Decided: arithmetic that exceeds its type (`value: int8 := 127; value + 1`) and converting a non-constant value that doesn't fit (`uint8(x)` with `x` = 256) stop the program with an error, like Swift, Zig's safe builds and Rust in debug builds. This departs from Go, which wraps silently. The evaluator already behaves this way, so nothing changes. The README documents it.

The specification already decides one runtime behaviour: reading a missing map key is an error ("If you try to access a property by an unset key you will recieve an undefined error"), unlike Go's zero value. That stays.

## Checks

### Keep: they depend on runtime values

| Check | Where |
| --- | --- |
| Array and slice index out of bounds, or outside the supported range | expressions.go:187–209 |
| Missing map key | expressions.go:230 |
| Division by a zero value | expressions.go:656 |
| Integer overflow in arithmetic and negation | `checkedIntegerResult`, evaluator.go:159 (traps, decision 2) |
| Converting a non-constant value that doesn't fit | expressions.go:706 (traps, decision 2) |

### Remove: integer coercion

| Code | Why it can go |
| --- | --- |
| `coerceRuntimeArgument` (evaluator.go:91) and every call to it: declarations, assignment, arguments, return values (`coerceFunctionResult`), array and slice elements (`evalTypedExpressions`), map keys and values, map index and index assignment, struct literal fields | Inference gives every literal its type, and the analyzer checks every value against its destination |
| Untyped operand handling in `evalIntegerInfixExpression` (expressions.go:619–635) | Both operands always have the same concrete type |
| The untyped fallback for integer literals (expressions.go:61–64) and the `resultType == nil` fallbacks | Inference always records a type, so these become internal errors |

### Remove: already guaranteed by earlier phases

| Check | Guaranteed by |
| --- | --- |
| `loopDepth`/`loopLabels` tracking; `break` or `continue` outside a loop, unknown labels (statements.go:317–347) | Resolver |
| "identifier not found" (expressions.go:344; statements.go:101, 151) | Resolver |
| Unknown struct, not a struct (statements.go:177, 181) | Resolver |
| Operator type mismatches, "unknown operator", negating unsigned (expressions.go:593–610, 672, 688, 725, 732) | Analyzer and inference |
| Integer conversion argument count, converting a non-integer (expressions.go:694, 702) | Inference |
| Wrong argument count, not a function, missing function value (expressions.go:238–269) | Analyzer |
| Tuple value and arity (statements.go:114–141) | Analyzer |
| Unknown field, field given twice, missing required field (statements.go:192–209) | Analyzer |
| Private field reads and writes, `hasStructAccess` (members.go:36, 185, 201) | Analyzer |
| Unknown members, type functions and methods; wrong argument counts for `append` and `toSlice`; "has no methods" (members.go:43–159) | Analyzer |
| `len()` argument count and kind (builtins.go:14, 25) | Analyzer |
| Index on a non-indexable value, unhashable map key, assigning into a non-map (expressions.go:123, 225; statements.go:248, 287, 302) | Analyzer |
| Invalid map type, invalid assignment target (statements.go:78, 228) | Parser and analyzer |

Where a check guards something that cannot happen, the code can either drop it or replace it with a short internal error so a bug in an earlier phase still fails loudly. Using `panic` for "should never happen" would be the clearest signal.

### Dead code

| Code | Why |
| --- | --- |
| Range over a string (expressions.go:553–562) | The analyzer only allows ranging over arrays, slices and maps |
| `isTruthy` treating non-booleans as true, `!` on non-booleans returning false | Conditions and `!` operands are always `bool` |
| Duplicated `if resultType == nil` (expressions.go:639–645) | Repeated twice in a row |

## Tests

`testEval` currently ignores errors from the earlier phases so it can feed the evaluator invalid programs. Once the evaluator trusts the program, `testEval` should fail the test on any phase error, so every evaluator test runs a valid program.

| Test | Change |
| --- | --- |
| `TestErrorHandling` | Delete. Every case (`5 + true`, `-true`, `true + false`, `foobar`) is now a resolver, inference or analyzer error, and those phases already test them |
| `TestLenBuiltin` error cases (`len(1)`, two arguments) | Delete the error cases; the analyzer tests cover them |
| `TestForLoops` cases for `break`/`continue` outside a loop and unknown labels | Delete; the resolver tests cover them |
| `TestIfElseExpr` case `if (1) { 10 }` | Delete; the analyzer covers it |
| `TestEvalRejectsInvalidStructFieldAccess` | Delete; the analyzer covers private reads and writes |
| `TestFunctionClosures` first case (calls through `function`) | Rewrite with a signature type, or expect the compile error, depending on decision 1 |
| `TestCheckedIntegerOverflow`, `TestCheckedRuntimeIntegerConversion` | Keep; they check the trapping behaviour from decision 2 |
| Everything else | Keep, running through the checked `testEval` |

## Order of work

1. Settle decision 1.
2. Make `testEval` fail on phase errors and delete or rewrite the tests listed above.
3. Remove integer coercion.
4. Remove the checks guaranteed by earlier phases, and the dead code.
5. Apply decision 1 in inference or the analyzer.
