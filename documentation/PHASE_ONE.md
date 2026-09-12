###### Instructions for AI

- Do not directly edit this file
- If you're going make direct suggestions, compare code against this document and the specification only
- Good enough is good enough, this is a solo project

# Phase One

Phase one was primarily focused on building out and proving the initial question poised; could i extend the Monkey language from the book Creating an Interpreter In Go. The result has largely been yes, and I consider it a successful first pass at doing so. In each of the major topics I will briefly outline some of the variations from the Monkey language, what I am happy with and what I feel could be done better for a second pass. This could be language or architectural changes.

### Lexer

The lexer has mostly been lifted whole from the monkey language with only minor additions. The largest addition to the lexer has been the inclusion of a line and column tracker. I'm not a big fan of the gigantic switch statement but it does allow me to use fine grained control of look ahead (= is assign, == is equals) for determining more complicated tokens.

I believe if I was going to try and do this differently it would be to perhaps create a map of single and double character tokens though I would have to research what effect or benefit this would have beyond modern sensibilities/dislike of switch statements.

### Token

One of the biggest changes from Monkey, though it might seem minor is how I have created the tokens. I had a few goals in mind which I believe I have achieved

- I have used iota with uint8 for token type compared to the string type in Monkey, a 16x reduction in memory used for the token type field
- We use a lookup table for the token names which is a one off memory cost, not a huge issue
- The largest update to the Token struct is the positional information of the token allowing for more comprehensive debugging and, down the line, better formatting

I believe some improvements are able to be had surrounding GetTokenName being needed and perhaps the Literal being assigned. In a lot of cases the literal will always be the name of the token so it might be worth investigating a solution that perhaps adds performance.

### Parser

The parser has changed by way of extensibility and I took an experiment with how the file structure is shaped. I have major features in their own files (e.g. collections_test.go, collections.go, functions_test.go, functions.go etc.) which I am not sure I like. I originally split them out into their own features due to the size of the language beyond Monkey but it is largely a similar mental overhead to parsing that many file names over searching through a file.

The switch usage is also a valid argument here as it is in the lexer. The files parser.go, statements.go, and expressions.go rely on switch statements to route the parser towards the correct abstract syntax tree type. This is fine for the most part but it can get confusing as to where your current syntax is. The flip side of this is that as the language becomes more complete, adding new features becomes genuinely easier.

On the whole, I am happy with my work in the parser and would like to consider the following future changes.

- Consolidating features into larger files (statements, expressions) over per feature files
- Adding a debug mode with event driven logging over the silent logging that currently takes place
- Improving the error tracking to be more consistent throughout the parser and project

### Semantic

The single biggest change to the project is the addition of the semantic analyzer. The semantic analyzer takes a fully formed ast and then, once again, parses through it to achieve two major goals

- Enforce type safety and error on any violates
- Infer types that are unknown

This system works largely like the parser and makes use of powerful utilities from the types system to compare and contrast what has been picked up.

One of the major sticking points of the semantic system is that it is now overloaded with responsibilities. When I made the decision to infer values at this step it also brought some (not all) of integer evaluation, which magnifies the complexity of the system altogether.

This has been a pain point for me as it is my least favourite place to develop in, as it is largely a repeat of the process I have completed in the parser and is, in some ways, far more complex.

A large solution to this would be an inference system placed between the parser and the semantic analyzer which resolves types and places initial checks on values. This would free the semantic analyzer to concentrate on the environment created, performing look ups to variables, ensuring boolean operators are valid and that everything the user is trying to achieve is within the rules of the language.

### Types

The types system was a big goal of mine for this project. I like JavaScript, hate TypeScript and enjoy working in Go, Rust, and Java. I believe by giving the user a real type system to work with I hand them the power to make their programs more safe, efficient, and defined.

The type system includes clearly defined types, along with their specific metadata. Builtins which are used by the semantic analyzer (and sometimes evaluator) to perform checks on user code and members.

The members are probably some of the more fun to code points within the Primordial. These are functions that can be used to extend the functionality of maps, arrays, slices etc. and move the language into that sweet spot between JavaScript and Rust.

I think it's too early in the types system to evaluate "how it's going" and we need to expand the systems around it before truly honing in on improvements. This is definitely one to watch, though.

If I had to consider one thing it would be to bring a lot of the integer control into this package as it is currently spread across the semantic analyzer, and the evaluator system, which is inefficient. This change would also improve the lowering layer between the parser and the semantic evaluator by simplifying it's role and controlling where decision making lays.

### Evaluator

Probably the weakest part of the proejct. This, again, was lifted wholesale out of the Monkey book with little to no thought for improvement. Part of this was because I was looking ahead to the compiler step, so I asked myself the question - why? Looking back I think it wouldn't have been too much effort to do some basic improvements such as introduce an interface/struct system as with the parser and semantic analyzer.

This would have allowed for more thought out error handing, grouping, and cleaner extensibility. I may still do this as it is perhaps a good place for an AI offload.

### REPL

Much like the evaluator, I have not looked into the REPL all that much, preferring to offload to the AI for quicker results as I focus in on the language structure. This currently allows me to run files from the command line though which is a very cool experience as the language starts to have that early feeling of 'being real'.

./pri run examples/server/server.pri

### Integers

The current integer system works with eight fixed with integers (int8 through uint64).

Typing would usually give an integer its concrete type. The context surrounding the integer gives the integer its concrete type (e.g. defined types on declaration, parameters, return types, etc.). Should no context be provided then the integer will default to int64 or uint64.

There is no silent adjustments to an integer, should you assign a negative value to an int8 it will not become a uint8. Instead an error will be thrown to inform the user of their mistake.

If a value is to overflow it will throw an error.

I would like to add in call like syntax to adjust a type for a variable or number.
