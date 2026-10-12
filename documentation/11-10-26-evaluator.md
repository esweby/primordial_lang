# Evaluator Analysis

The evaluator currently solves three problems at the same time, this has led to it becoming overloaded and feeling unwieldly to work in. The three problems are:

- Creation of the environment to tack the current scope and retrieve symbols available to it
- Inferring undeclared types
- Analysis that all use of types and values are valid within the boundaries of the language

It also solves a secondary problem, namely, the definition of integers within the language. This is less than ideal as inferring exact integer types and working with them being wound into the workings of the analyzer has proved to be a tough problem that makes understanding the individual sub tasks weighty and makes adjusting language rules difficult.

## Solution

Create a layer for each of the three tasks the analyzer does and give them a defined scope.

1. Resolver - creates a scope that can be followed from start to end and used within the subsequent phases for accurate inference and analysis
2. Inference - ensures that program nodes are accurately typed with context are that integers are managed appropriately
3. Analyzer - sole duty is to make sure that with the statements and meta data provided, the program follows the rules of the language

This will be a staged approach where the first two phases will be created in turn, then the analyzer will be slimmed down to fit it's new responsibilities.

### Resolver

The resolver package will be responsible for building out an accurate environment and enriching the existing ast with the associated meta data. Within the AST we will need to adjust several nodes (identifiers, mostly) to point to the declaration node as it is resolved within this phase. The resolver itself will largely follow the same pattern as the existing symbol table within the evaluator and the subsequent packages will perform their functions directly on the program, as before. This means that when they come to an identifier they will be able to directly access the referenced node as it has already been resolved.

This purposefully simplistic approach, while dilutiing the ast, also enriches it with more metadata and avoids any over architecturing at this point of the languages lifecycle.

### Inference

The inference phase will be responsible for assigning a specific type where none is provided by the user. It will also be responsible for ensuring that integer types are concretely assigned within the parameters of the languages rules.

Integers follow Go's pattern. An integer literal is an untyped constant until it meets a type, and that type comes only from the statement it appears in, never from later uses of the variable. Once a variable has a type it is fixed.

- An untyped constant takes the type of whatever it is used with: a declared type, the other side of an operator, a parameter, a return type, or a collection's element type.
- If there is no such context the constant defaults to `int`.
- A constant that does not fit the type it is given is an error, e.g. `x: int8 := 300;`.
- `int` and `uint` are their own 64-bit types, distinct from `int64` and `uint64`.
- There are no implicit conversions between integer types. Mixing them requires an explicit conversion such as `int32(x)`.
- `len()` and range indexes produce `int`.

In the following example x is an `int`, so adding it to an `int32` is an error.

```
x := 0;            // x is int
y: int32 := 10;    // 10 takes the declared type int32

z := x + y;        // error: mismatched types int and int32
a := int32(x) + y; // a is int32
w := y + 1;        // 1 takes the type of y, so w is int32
```

### Analyzer

The analyzer will then be left with the primary responsibility of validating the program types as it already does. There is very little to write about this as it will no longer be taking language actions but enforcing language decisions and in a cleaner, more compact package.
