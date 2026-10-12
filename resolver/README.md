Note: All code examples are loose pseudo code, not definitive data structures and not absolutely set in stone.

# Resolver Phase

The resolver phase will create an interactable

The resolver phases builds an environment object that can be used in subsequent analysis phases. Moving the environment building out of subsequent analysis phases removes unrelated code from the subsequent phases and allows the language to more easily expand with more complex ideas in the future. An example of a more complex idea would be a deterministic movement control for the user, the ability for a user to move the scope of a variable into a calling function for earlier cleanup. While unneeded for the moment, weaving that through the semantic analysis phase as is would prove to be hacky and may introduce bugs, with these clean phases in place we can explore ideas like that.

### Scope & Symbols

The scope data structure will broadly follow the pattern of the previous symbol table except for one major change, it will allow for two way traversal. An additional benefit could be the inclusion of a map[string]Symbol for o1 lookups of symbols by name when available in the current scope.

```Go
Scope {
  symbols []Symbol
  parent Scope
}
```

The symbol data structure will also be similar to the previously implemented symbol. The primary difference will be that a concrete type will be recorded on the associated node, not on the symbol directly.

```Go
Symbol {
  Name string
  Node *ast.Node
  ParentScope *Scope
  Scope *Scope
}
```

It's important to note that while all symbols may have their own scope, not all scopes will have their own symbol. An if branch might be an example of this or a block statement being assigned to a variable would be another example of a scope being created.

## Changes to the AST

As the inference and semantic phases will still need to work from the program and evaluate statements given to come to the correct version, each ast node that can hold a scope will need to have a scope included in it. To defeat circular logic at the top level of the ast will be an interface that mimics (lightly) the scope and symbol from the resolver package. These would then be placed as \*pointers on the relevant nodes.

The interfaces would follow this pattern:

```Go
Scope {
  ParentScope() Scope
  FindInScope(name string) Symbol
}

Symbol {
  Name() string
  Type() types.Type
}
```

A non exhaustive list would be

- Symbol nodes
  - declare statement
  - identifier
  - function declarations
  - tuples

- Scopes
  - if statements
  - block expressions
