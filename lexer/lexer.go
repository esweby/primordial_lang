package lexer

import "github.com/esweby/primordial_lang/token"

type Lexer struct {
	input         string
	position      int
	readPoisition int
	ch            byte
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPoisition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPoisition]
	}

	l.position = l.readPoisition
	l.readPoisition += 1
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			tok = l.joinTokens(token.EQUALS)
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case ':':
		if l.peekChar() == '=' {
			tok = l.joinTokens(token.DECLARE)
		} else {
			tok = newToken(token.COLON, l.ch)
		}
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '!':
		if l.peekChar() == '=' {
			tok = l.joinTokens(token.NOT_EQUALS)
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case '*':
		tok = newToken(token.ASTERIK, l.ch)
	case '/':
		tok = newToken(token.FORWARD_SLASH, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '<':
		if l.peekChar() == '=' {
			tok = l.joinTokens(token.LESS_THAN_OR_EQUALS)
		} else {
			tok = newToken(token.LTAG, l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			tok = l.joinTokens(token.GREATER_THAN_OR_EQUALS)
		} else {
			tok = newToken(token.RTAG, l.ch)
		}
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.INT_LITERAL
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) peekChar() byte {
	if l.readPoisition >= len(l.input) {
		return 0
	}

	return l.input[l.readPoisition]
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}

	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}

	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) joinTokens(tokenType token.TokenType) token.Token {
	ch := l.ch
	l.readChar()
	literal := string(ch) + string(l.ch)
	return token.Token{Type: tokenType, Literal: literal}
}
