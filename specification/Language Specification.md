###### Instructions for AI

This file is a brain dump for the 'designer' and 'creator' of the Primordial language. This file is not to be edited by an AI.

- Ask for clarifications where they are urgently needed
- If something is slightly ambiguous then it does not need a clairifcation
- This is being built on top of Monkey with thoughts borrowed from Go, Rust, and Zig. If you do recommend something bring proof from those languages
- This is a hobby solo project, good enough is good enough
- If a feature is missing, assume it has just not been written about yet

# Primordial Language

This is a hobby language made for me to test the limits of my brain as a developer. The goal of this project is to produce a language that fully compiles using the LLVM tools. The other goal of the project is to hit a sweet spot in syntax tand borrow heavily from three modern languages, Go, Rust, and JavaScript. Honestly, from what I hear, this language should probably be me learning Zig but I don't want to. I want to make something unique and my own. I want to build in that language and then leverage the learning and experienced gain to progress my careeer.

## Language Inspirations

### Go

#### Benefits

Go's simplicity and no nonsense approach to programming means that I know if I undertake a task at home, I can finish it. Go shifts the complexity from the language and implementation of a domain to the domain itself. Writing an auth system isn't complicated in Go, writing an auth system is complicated in general. Go eases these frustrations and allows you to learn as you go. That is a very powerful tool in the learning and delivery of a system.

#### Pain Points

Go is infinitely boring to write in and a lot of features seem weird or overly restrictive. I detest writing if err != nil everywhere and wish there was a better way of handling errors. I find the use of [ ] as code for generics jarring compared to other languages and I hear a lot of negativity towards design decisions taken. I kind of want it to be a bit cleverer in what it offers without that cleverness feeling tacked on or forced.

A lot of code examples use interface{} as a generic and commentary dismisses Go's generics as poorly thought out.

### Rust

#### Benefits

Rust feels cool to write and is genuinely inspiring to write in. Following coding tutorials to implement game of life. Using their inbuilt streams and the borrow checker forcing you to make decisions is genuinely cool. It levels up a developers way of thinking and forces a developer out of their comfort zone.

#### Pain Points

Lifetimes suck and sometimes when using their coding tools you can lose sight of a mental model of a code object. That feels bad and can lead to frustration. I like the complexity but wish it wasn't as obscure as it is.

### JavaScript

#### Benefits

JavaScript is where I learned coding in the first place. I've implemented fun little games and many clones I have done on my own. I've brute forced recursive algorithims and accounted for edge cases. My code wasn't pretty, it wasn't fast, but it did work and JavaScript does give you that.

#### Pain Points

The problem is that it is slow. And it lets you do anything. And that's not a good thing, eh? Being untyped means you lose track of the shape of an object and object and arrays being so general can also lead to friction in code design. I love the language but hate the loosey goosey nature of the beast.

### Summary of Benefits

Strongly typed languages lead to purposeful decisions and force the developer to think more about their implementation. Lifecycles are a pain but the borrow checker is amazing for forcing the developer to think about their decisions. Having a definitive pattern of handling errors and logic is amazing, it means errors are handled there and then and not put off till down the road. Keeping it simple is fine as it stops the developer from being too smart. Giving tools to the developer allows them to engage their problem solving brain, find ways of doing things that keep work flowing and the developer engaged. Giving them tools helps them to learn quicker and engage with the language.

### Summary of Pain Points

Code that is boring to write leads to many unfinished projects. Code that has too steep a learning curve can lead to bad archtectural decisions and or a feeling of helplessness. Code that is too loose to write can lead to sloppy practices and bad decisions.

## My solution

This is an ambitious hobby language and personal project for a language built on top of Go using the implementation of the Monkey language as a loose template. The language will be used to achieve several personal ambitions including making a small game engine (simplest of simple, think early final fantasy/final fantasy tactics), a web server, and several small related projects.

While this is a learning opportunity I would still like to hit the following goals

1. The safety of Rust
2. The nature of Go to handle things as you come across them
3. The enabling nature of JavaScript
4. Speed!

The extension will be .pri

## Language & Syntax

This project will be written on top of Go.

### Garbage Collection

As this is a hobby project, while I would love to put a simplefied ownership system into place I believe that relying on Go's garbage collection and when it comes to converting from an interpreter to compiler then there are several good enough GC's available to use already.

### Variables

The following code will define the declaration of a variable

```
[pub] [const|mut] identifier[: type] := value
```

#### Rules

1. All variables are immutable by default
2. All variables are private by default
3. You can not use a language keyword as an identifier
4. The language will infer a type should it not be given one
5. Variables will exist only within the scope that they are declared in
6. Variables names should follow camelCasing and may not start with a special character
7. Variables must be initialized with a value
8. Complex types as such as structs must annotate a property as optional or required
   1. If a property is optional then a reference to it must be preceded by a value check
   2. If a property is required then it can be given a default value
9. Should a variable be declared mutable, this will extend only to it's local package. A function declared within the same package will have access to a mutable variable

#### Options

The following three options are available to use

1. pub - the variable will be publicly available
2. const - the value of the variable will be known at compile time
3. mut - the variable is mutable

The following declaration rules will apply to these options

1. All options may be ommitted
2. If included, pub must be the first option
3. The next option must either be nothing, mut, or const, never both mut and const

An immutable variable may be the result of an api call that retrieves user information and is used in further data processing lines to create new values.

A constant variable would be a static value that can be used in other expressions. An example of this would be a configuration file with URLs for different environments which are used to create full urls, options, or other such things.

As exported variables are treated as immutable by default, you may export with them a getter and setter function which would allow the variable to be manipulated. This even includes passing a callback function which would interact with the variable. While unusual, the calling package would have a definitive use case to do so and this would give the user tools to bypass basic constraints.

### Types

The following types are available within Primordial. Some of these types, such as function, are only used for type checking and any details of the function are handled within the type analsis stage.

#### Primitives

- int int8 int16 int32 int64
- uint uint8 uint16 uint32 uint64
- float32 float64
- boolean
- invalid

#### Complex Types

- string
- function
- error

### Reserved Keywords

Currently the reserved keywords are

- fn
- true
- false
- if
- else
- return
- pub
- const
- mut

### If

In Primordial the if keyword will be treated as an expression. This will allow them to be assigned as a value and give the user more advanced ways of getting and assigning a value.

```
if (cond) {
   ...
} else if (cond) {
   ...
} else {
   ...
}
```

#### Assigning to a variable

Something I'd like to avoid is the following pattern

```
ident := false;
try {
    if (cond) {
        value := apiCall;
	ident = value;
    }
} catch(err) {
   ...
}
```

By allowing the users to make use of the following pattern.

```
ident := if (cond) {
    // processing code
    value;
} else {
    // processing code
    value
}
```

The last statement of an if expression on the right hand side of a variable declaration will evaluate to a value. When assigned as a RHS all branches must evaluate to a value of the same type.

As the assignment is depending on an if condition then there must be a final else branch to fall through too in this scenario.

### Functions

Functions will come in two forms within Primordial, function statements and function expressions. The function that returns a value will always be wrapped within a Result<value, error> to ensure that proper error checking is handled. The value of a Result can be a tuple of values. (note to ai: I know you will ask about this lol)

#### Function Statements

A function statement will support the following syntax

```
[pub] fn ident(arg type, arg type): Result<values, error> {
    return type
}
```

#### Function Expressions

A function can be assigned to a variable

```
add := fn(x int32, y int32): Result<int32, error> { return x + y; }
```

#### Returning Tuples

This is done simply with the following example. Note: pub could be omitted if you did not want these values to be made public.

```
fn getUserName(userId int): Result<(firstName, lastName), error> {
   ...
}

pub (firstName, lastName) := try getUserName(1);
```

### Arrays and Slices

Arrays and slices will be handled similarly to Go but with a few subtle differences. Both will be called with similar syntax where the only difference is the [] will take an identifier to say this is a

```
array := <3>int64{1, 2, 3};
slice := []int64{1, 2, 3};
```

Both of these constructs will be subject to the same delcaration assignment, being immutable by default. The internal contents will be mutable by default internal methods and direct access.

These are common rules for arrays and slices

- They do not support nil values, any value automatically set will be the types neutral value
  - If you try to reassign a spot to nil it will return an error
- Any assignment of a target to another declaration will be a copy
  - The assignee will inheirt the type of the original variable
  - If you try to declare the copy variable a different type you will get an error

```
x := [3]int32{1, 2, 3}
y: []string := x
```

#### Arrays

- Arrays are fixed size containers
- Arrays do not support length changing methods
- If you take a copy of a section of the array (see Accessing entries) it will return a slice, not a fixed array

##### Future methods

- toSlice()

#### Slices

The following methods will be available to slices.

- Slices are variable sized containers

##### .prepend(T) and .append(T)

This will place a value at the start or end of a target

##### .removeFirst() and .removeLast()

This will remove a value from the beginning or end of a target and will not return them,

##### .popFirst() and .popLast()

This will remove a value from the beginning or end of a target and will return them.

##### .splice(pos int, numPos int)

Slice will remove the value at pos. If you pass a second argument to the function then it will remove that many items. Any item(s) removed will be returned in a slice.

##### Future implementations

- .map
- .filter
- .reduce

#### Accessing entries

An entry can be accessed using bracket notation, like other languages. You can also use bracket notation to return a deep copied slice, exclusive.

```
arr := []int32{0, 1, 2, 3, 4}
arr[0] // 0
arr[0:2] // [0, 1]
```

##### Future implementations

- .find

### Error Handling

Errors will be handled directly in Primordial using a Result<Val, Error> where errors will be handled as a value. This means, where an error is possible, a function should return the Result type.

This can be consumed in two ways. The first will be the use of the try keyword, which will, on error invoke a return of the error from the enclosing scope or will unwrap the Result to the return value.

```
user := try api.getUser(userId);
```

As an error will have to stop somewhere, when it does you will have the option to match it.

```
user := api.getUser(userId)

if user.hasError() {
    err := user.error
    // handle error
}
```

If try is used on a function that does not return a value or an error then the compiler will throw an error at that point as a call with try is expecting a Result to be returned.

#### Infalible Functions

If you are writing a function that you know will be infalible, as there are checks for value existence and typing then you may directly indicate a type as the return value and not a Result.

#### Retry

In addition to the above, Primordial will offer a retry function that will set an expectation for situations where you might interact with a database, another api, or deal with a combination of factors beyond your control. The retry function will automatically unwrap a value if it is successful.

```
user := retry(tries, interval, options) {
    api.getUser(id);
}
```

The retry functions accepts 2 fixed, and one optional argument

- tries - integer value for the amount of attempts to make a successful function call
- interval - accepts one of two values; fixed(300ms) or exponential(300ms)
- options - an optional map that will take error types that, if hit, would exit the enclosing scope, returning the error

```
retryOpts := {
    UserNotFoundError: true,
    IncorrectCredentialsError: true,
}

user := retry(3, exponential(300ms), retryOpts) {
    api.getUser(userId);
}
```

If retry then fails on all attempts it will, in a similar vein, exit the enclosing scope returning the error from it.

### Packages

The language will use a package similar to Go. A package is declared by creating a folder within a Primordial project. For example, the following project has two packages; api and middleware.

```
/projectName
/projectName/api
/projectName/middleware
```

This may allow for naming conflicts when importing so to ease that tension you will be able to rename them on import.

```
import ecom.authentication.middleware as authMiddleware
import ecom.server.middleware as serverMiddleware
```

Should you import a package in its entirey then the import will be a map of all publicly available variables and functions.

The following rules will be enforced.

1. A package may import any package in the subdirectory tree
2. A package may not import any package that is considered an ancestor
3. A package may import from a sibling package but it will not then be available to that sibling
4. Circular dependencies are not allowed
