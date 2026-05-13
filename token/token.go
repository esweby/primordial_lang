package token

type TokenType uint8

const (
	ILLEGAL TokenType = iota
	EOF

	// Identifiers
	IDENT
	INT_LITERAL
	STRING_LITERAL

	// Types
	INT8
	UINT8
	INT32
	UINT32
	INT64
	UINT64
	FLOAT32
	FLOAT64

	// Operators
	ASSIGN  // = for assigning a value
	DECLARE // := for assigning variables
	COLON
	PLUS
	MINUS
	BANG
	ASTERIK
	FORWARD_SLASH
	EQUALS
	NOT_EQUALS
	GREATER_THAN_OR_EQUALS
	LESS_THAN_OR_EQUALS

	// Delimeters
	COMMA
	SEMICOLON

	LPAREN // (
	RPAREN // )
	LBRACE // {
	RBRACE // }
	LTAG   // <
	RTAG   // >

	// Keywords
	FN
	TRUE
	FALSE
	IF
	ELSE
	ELSEIF
	RETURN
	PUB
	MUT
	CONST
)

var tokenNames = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	// identifiers
	IDENT:          "IDENT",
	INT_LITERAL:    "INT_LITERAL",
	STRING_LITERAL: "STRING_LITERAL",
	// Operators
	ASSIGN:                 "ASSIGN",  // =
	DECLARE:                "DECLARE", // :=
	COLON:                  "COLON",   // :
	PLUS:                   "PLUS",
	MINUS:                  "MINUS",
	BANG:                   "BANG",
	ASTERIK:                "ASTERIK",
	FORWARD_SLASH:          "FORWARD_SLASH",
	EQUALS:                 "EQUALS",
	NOT_EQUALS:             "NOT_EQUALS",
	GREATER_THAN_OR_EQUALS: "GREATER_THAN_OR_EQUALS",
	LESS_THAN_OR_EQUALS:    "LESS_THAN_OR_EQUALS",
	// Delimeters
	COMMA:     "COMMA",
	SEMICOLON: "SEMICOLON",
	LPAREN:    "LPAREN", // ()
	RPAREN:    "RPAREN",
	LBRACE:    "LBRACE", // {}
	RBRACE:    "RBRACE",
	LTAG:      "LTAG", // <>
	RTAG:      "RTAG",
	// Keywords
	FN:     "FN",
	TRUE:   "TRUE",
	FALSE:  "FALSE",
	IF:     "IF",
	ELSE:   "ELSE",
	RETURN: "RETURN",
	PUB:    "PUB",
	MUT:    "MUT",
	CONST:  "CONST",
}

var keywords = map[string]TokenType{
	"fn":     FN,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"pub":    PUB,
	"mut":    MUT,
	"const":  CONST,
}

type Token struct {
	Type    TokenType
	Literal string
	Line int
	Column int
}

func GetTokenName(tokenType int) string {
	return tokenNames[tokenType]
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}

	return IDENT
}
