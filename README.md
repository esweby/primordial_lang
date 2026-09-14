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

## Syntax at a glance

The following syntactical features have been implemented

### Variable Declaration

Variables are immutable by default, always typed either by annotation or inference.

```
x := 123;
y: int32 := 123;

mut z: string := "Hello world";
```

### Functions

Functions can be declarations or assigned as values.

```
fn add(x int32, y int32): int32 {
	return x + y;
}

add := fn(x int32, y int32): int32 { return x + y; }
```

### If expressions

Primordial supports regular if statements but also if branching as expressions

```
age := 19;
mut canDrink := false;

if (age >= 18) {
	canDrink = true;
}

// This can become the following

canDrink := if (age >= 18) {
	true
} else {
	false
}
```

### Collections

It also supports arrays, slices, and maps using Go style syntax. In an array with a defined length, if you would provide less values than the size of the array, the remaining places will be initialized to the 0 value.

```
// Array
ages := [3]int32{ 24, 26, 18 };

// Slice
names := []string{ "Graham", "Francis", "Ethel" };

// Maps
ages := map[string]int32{
	"graham": 24,
    "francis": 26,
    "ethel": 18,
};
```

### For loops

There are four different styles of for loops to choose from, which support break and continue keywords.

```
// infinite
for {
	...
}

// for while
mut age := 17;
for (age < 20) {
	age = age + 1;
}

// traditional for (the initializer does not need to be marked as mut)
for (x := 0; x < 10; x = x + 1) {
	...
}

// range over collection
ages := []int32{ 10, 20, 30, 40 };
for i, age := range ages {
	age;
}
```

### Structs

And finally struct implementations.

```
struct Person {
	age: int32; // internally mutable only
    pub name: string; // externally mutable

    // static methods
    fn new(age int32, name: string): Person {
    	// supports short form
        return Person{
        	age,
            name,
        }
    }

    // impl block is for initialized methods
    // and makes the self keyword available
    impl {
    	fn setAge(age int32) {
        	self.age = age;
        }

        fn getAge(): int32 { return self.age; }
    }
}

tobias := Person.new(400, "Tobias");

tobias.setAge(40);
tobias.getAge(); // 40

tobias.name; // "Tobias"
tobias.name = "Tobi";
```

## Give it a try

Use the following commands to give the language a try

```sh
go build -o pri main.go
./pri run examples/main.pri
```

Run the test suite:

```sh
go test ./...
```

There is a more structured [language specification](specification/Language%20Specification.md). Both are working documents, not sacred texts.

Implementation documentation and retrospectives lives in [`documentation`](documentation/README.md).

This is an ambitious solo hobby project. Good enough is good enough—until it becomes interesting to make it better.
