package token

type TokenType uint8

const (
	ILLEGAL TokenType = iota
	EOF

	// Identifiers
	IDENT
	INT_LITERAL

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

	// Delimeters
	COMMA
	SEMICOLON

	LPAREN // (
	RPAREN // )
	LBRACE // {
	RBRACE //}
	LTAG   // <
	RTAG   // >

	// Keywords
	FN
	TRUE
	FALSE
	IF
	ELSE
	RETURN
	EXP
	MUT
)

var tokenNames = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	// identifiers
	IDENT:       "IDENT",
	INT_LITERAL: "INT_LITERAL",
	// Operators
	ASSIGN:        "ASSIGN", // =
	DECLARE:       "DECLARE", // :=
	COLON:         "COLON", // :
	PLUS:          "PLUS",
	MINUS:         "MINUS",
	BANG:          "BANG",
	ASTERIK:       "ASTERIK",
	FORWARD_SLASH: "FORWARD_SLASH",
	EQUALS:        "EQUALS",
	NOT_EQUALS:    "NOT_EQUALS",
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
}

var keywords = map[string]TokenType{
	"fn":     FN,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"exp":    EXP,
	"mut":    MUT,
	"int":    INT_LITERAL,
}

type Token struct {
	Type    TokenType
	Literal string
}

func (t *Token) ToString() string {
	return tokenNames[t.Type]
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}

	return IDENT
}
