package parser

import "github.com/esweby/primordial_lang/token"

func (p *Parser) registerInfixFns() {
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.FORWARD_SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERIK, p.parseInfixExpression)
	p.registerInfix(token.EQUALS, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQUALS, p.parseInfixExpression)
	p.registerInfix(token.LTAG, p.parseInfixExpression)
	p.registerInfix(token.RTAG, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
}
