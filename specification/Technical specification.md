##### Instructions for AI

Do not overwrite the contents of this docment. This document is for editing by the repo owner, only. Please read this document to understand the implementation of the custom programming language.

# Primordial

Extension: .pri

An ambitious hobby language built on top of Go as a learning experience and to fulfil several personal ambitions. The language will be used to build anything from a game engine, to a web server, and connecting projects inbetween.

While the goal of the language is to learn and level up as a programmer, I would like to hit the following goals:

1. The safety of Rust and Go
2. Simplicity of Go
3. The enabling nature of JavaScript to get things done
4. Fast!

## Packages

This language will use a package based system similar to Go with a few key differences. A package is simply declared by creating a folder within a Primordial project. For exmaple, the following project has two packages: api and middleware.

```
/projectName
/projectName/api
/projectName/middleware
```

You would then be able to import from another package using two keywords, **import** and **as**. Due to the nature of how you declare packages there will be times where there are naming conflicts, this will be avoided using the as keyword.

In the below example, the project ecom has two folders in its base directory: authentication, and server which create the package structure. Both of those packages and the children of both, named middleware, are then available.

```
#import ecom.authentication.middleware as authMiddleware
#import ecom.server.middleware as serverMiddleware
```

Imports will always follow the following pattern projectRootFolder.folderName.folderOrPropertyName. In the case that you have a publicly available variable or function that is the same name as a child package, the publicly available variable/function will be prioritized first but you will receive a warning to highlight the bad naming strategy.

Should you import a package in it's entirely then the import will be a map including all of the publicly available variables and methods.

This has the possibility to get complicated so the following rules are enforced

1. A package may import any package in its subdirectory tree
2. A package may not import any package that is its ancestor
3. A package may import from a sibling package but it will not then be available to that sibling
4. Circular dependencies are not allowed

Should you find yourself using the same folder names across your project it should be for commonly used assets or groupings of functionality i.e. middleware, assets, components.

## Variables

A variable is declared using the := assignment syntax and has the following rules.

1. All variables are immutable by default
2. All variables are private by default
3. You can not use a language keyword as an identifier
4. The language will infer a type should it not be given one
5. Variables will exist only within the scope that they are declared in
6. Variables names should follow camelCasing and may not start with a special character
7. Variables must be initialized with a value
8. Complex types such as structs must annotate a property as optional or required
   1. If a property is optional then a reference to it must be preceded by a value check
   2. If a property is required then it can be given a default value

```
identifier := value
```

Variables will infer a type, where possible and most basic types available in Go will be initially available in Primordial.

- string
- int int8 int16 int32 int64
- uint uint8 uint16 uint32 uint64
- float32 float64
- boolean

All variables can be assigned options using < > syntax. The following three properties may be assigned, in any order, or any quantity. Each option will be separated by a comma. A mutable variable may be made publicly available, but may not be mutated outside of it's origin package.

- mut - is a mutable variable
- pub - is publicly available and able to be imported from other packages
- type - the type of the variable from any of the above types
- generic -

```
<mut> age := 17
age = age + 1

<mut, int32, pub> age := 32 // #import project.package.age
```

Mutable variables that are made publicly available are treated as read only when imported.

```
- project/user
<mut, pub> name := "Tobias"

- project/processing
#import project.user.name

name = "Cats" // Will throw an error
```
