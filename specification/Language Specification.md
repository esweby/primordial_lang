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
- map
- struct
- range
- for
- break
- continue

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

#### Collection Types

- tuple
- array
- slice
- struct
- map

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
array := [3]int64{1, 2, 3};
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

An entry can be accessed using an index expression (bracket notation), like other languages. You can also use bracket notation to return a deep copied slice, exclusive.

```
arr := []int32{0, 1, 2, 3, 4}
arr[0] // 0
arr[0:2] // [0, 1]
```

##### Future implementations

- .find
- deep copy

### Structs

Structs are Primordial's primary way of providing a logical grouping and encapsulation of proprties and methods. This will be borrowing heavily from Rust way of working with some syntactic flavour to make the process easier to use.

A struct will be declared using the following syntax

```
struct Identifier {}
```

#### Initialisation

Initialisation will be done by the user declaring a New method on the struct and returning a instance of that struct. While the language is young I won't be handling references too explicitly and will leave this a little naive for the moment.

```
struct Person {
	age: int32;
    name: string;

	fn new(age: int32, name: string): Person {
    	return Person{
        	age,
            name
        }
    }
}

person1 := Person.new(25, "Tobias");
person2 := Person.new(23, "Edward");
```

#### Properties

Properties can be declared and accessed on a struct but follow the same rules as regular variables, they must use the pub syntax.

```
struct Person {
	pub mut name: string;
    age: int32;

    fn new(name: string, age: int32): Person {
    	return Person{
        	name,
            age,
        }
    }

    impl {
    	fn setAge(newAge int32) { self.age = newAge; }
        fn getAge(): int32 { return self.age; }
    }
}

tobias := Person.new("Tobias", "29");
tobias.name; // accessible
tobias.age // is private to the struct
```

1. All properties are internally mutable from an impl block using the self.propery = assignment
2. For a property to be publicly accessible, it must be explicitly declared with the pub keyword
   1. Declaring a property as publicly accessible will also make that property publicly settable
3. If a property is considered private, it is internally availabkle to methods declared on Person and can be returned via that.

#### Methods

There are two distinct ways of declaring a method on a struct, within a impl container denoting that those methods have access to the self property and can only be used on an instance of the struct. Methods declared outside of the impl block are considered static methods and may be called as Person.new(), for exampl).

```
struct Person {
	name: string;
    age: int32;
    mut email: string;

    fn new(name: string, age: int32, email: string): Person {
    	return Person{
        	name,
            age,
            email
        }
    }

    impl {
    	fn getName(): string { return self.name; }

    	fn getAge(): int32 { return self.age; }

    	fn setEmail(newEmail: string): boolean {
        	self.email = newEmail;
        	return true;
    	}

    	fn getEmail(): string { return self.email; }
    }

}

tobias := Person.new("tobias", 29, "tobias@domain.com");
name := tobias.getName();
tobias.setEmail("new_email@domain.com");
email := tobias.getEmail();
```

The self { } code block means that all functions declared within that block are passed self implicitly.

### Maps

Maps in Primordial will be more or less lifted directly from Go lang as an explicit data structure that is easy to understand and use.

A full instantiation of a map:

```
map[int32]string{} // without starting values
map[int32]string{ 1: "two", } // with starting values
```

All properties within a map will be mutable by default.

```
x := map[int32]string{}
x[1] = "one";
x[1] // "one"
```

Currently accepted key types for maps are

- string
- any integer
- boolean

If you try to access a property by an unset key you will recieve an undefined error.

#### Iteration order

I am currently undecied on how I would like to iterate over a map. We may store them in an order by default but that feels slightly unintentional. We may lean on top of Go's language for the moment and have them intentionally out of order.

#### Map Ideas

Perhaps to handle the idea of maps being out of order we may look at another keyword that can be used with maps such as ordered or orderedMap which could hold a property ascending, true by default (so 1,2,3 or a,b,c etc.) and would hold an order for an entry. The property could be accessible by a builtin member expression to handle the hidden property.

### Loops

I would like looping to be simple by taking advantage of tuples and commonly known keywords. As this language already likes to make use of expressions to give your code semantic meaning I feel it is appropriate to lean on what Rust and Zig offer while maintaining the simplicity of go. There is no need for all of the keywords (for, while, do while) and will overload the for keyword to lean on the different types of looping that a user might look to do.

Another factor will be to consider how to handle labels. Part of me likes including a label with a \# or another label designation for use with break and continue but they can often make code look clunky and unclean. Another option would be to take the index key used, usually associated with one speciifc loop and overload that with a label. The problem here is that this moves away from the simplistic nature of the language and towards jargon, which is not ideal.

For now we will explore some basic options and revisit looping as the language expands.

Rules for looping

- Must include a way to break out of the loop
  - In a iterator loop that is handled for the user
  - In a conditional loop the codition must meet some basic sense tests
    - To be defined

The following will be the overall layout of a for loop

```
[label:] for
	[(cond)] // 1
	[(init; cond; increment)] // 2
    [tuple := range incrementer] // 3
{

}
```

- A label is optional
- You may choose to have either a while style for loop, a traditional for loop, or a go style range loop
- In Primordial for loops are expressions

#### for

The basic for loop will is a traditional infinite loop

```
for {
    if (x == y) {
    	break;
    }
}
```

#### for while

The for loop with a simple condition replaces a while loop in other langues

```
x := 0;
y := 10;

for (x < y) {
	x++
}
```

#### for with condition

These loops are common in every language.

```
for (x := 0; x < 10; x++) {}

x := 0;
for (x < 10) {
	x++;
}
```

#### for x := range incrementer

The range keyword can be used with an array, slice, and map. The range keyword will return a tuple which the user can choose to be destructured or not.

- Array & Slice - index, value
- Map - key, value - this will be as they're added and not in a particular order

```
arr := [3]string{ "a", "b", "c", };

for i, val := range arr {
	...
}

m := map[string]string{
	"a": "one",
    "b": "two",
    "c": "three",
}

for k, v := range m {
	...
}
```

### Labels

Labels will be a day 1 feature so that advanced user control flow is available for breaking nested loops as needed.

```
breaker: for {
	for (x := 0; x < 10; x++) {
    	if (x == 5) {
        	break breaker;
        }
    }
}
```

### Continue

The continue keyword will allow a iteration in a loop to skip to the next iteration.

```
for {
	if (x == 5) {
    	continue;
    }
}
```

### Break

The break keyword will be used to break out of a loop, break out of a specific loop, and/or return a value. When a loop is assigned as a value to a variable then a break keyword must be used with a value assigned to it.

```
// syntax for break
break;
break label;
break (value);
break label (value);

for {
	if (userInput == "quit") {
    	break;
    }
}

x := for {
	break (1 + 1);
}

breaker: for {
	if (true) {
    	break breaker;
    }
}

b := breaker: for {
	for(x := 0; x < 10; x++) {
    	if (x == 5) {
        	break breaker (5);
        }
    }
}
```

- If you use a for loop as a right hand side of a return or declaration then you must include a break with an expression
- If you use a return within this it will break the loop
  - If you are in a function scope it will immediately exit the function scope with the provide values (as defined by the function)
- As appropriate break accepts up to 2 values
- If a break returns a value then it must return the same type in every

### Error Handling

**Error Handling is marked as a second pass feature**

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

**Packages is marked as a second pass feature**
