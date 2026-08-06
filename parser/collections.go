package parser

import (
	"strconv"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/token"
	"github.com/esweby/primordial_lang/types"
)

func (p *Parser) parseLeftBracket() ast.Expression {
	tok := p.curToken
	length := -1

	p.nextToken()
	if p.curTokenIs(token.INT_LITERAL) {
		parsedInt, err := strconv.ParseInt(p.curToken.Literal, 10, 0)
		if err != nil {
			return nil
		}

		length = int(parsedInt)
		if length <= 0 {
			return nil
		}

		p.nextToken()
	}

	if !p.curTokenIs(token.RBRACKET) {
		return nil
	}

	p.nextToken()
	collectionType, ok := types.GetBuiltin(p.curToken.Literal)
	if !ok {
		return nil
	}

	p.nextToken()
	if !p.curTokenIs(token.LBRACE) {
		return nil
	}

	p.nextToken()
	contents := []ast.Expression{}
	if p.curTokenIs(token.RBRACE) {
		return p.returnCollection(tok, length, collectionType, contents)
	}

	for {
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			return nil
		}

		contents = append(contents, expr)
		if !p.peekTokenIs(token.COMMA) {
			break
		}
		p.nextToken()
		p.nextToken()
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return p.returnCollection(tok, length, collectionType, contents)
}

func (p *Parser) returnCollection(
	tok token.Token,
	length int,
	collectionType types.Type,
	contents []ast.Expression,
) ast.Expression {
	if length > -1 {
		return &ast.ArrayLiteral{
			Token:    tok,
			Size:     length,
			Type:     collectionType,
			Elements: contents,
		}
	}

	return &ast.SliceLiteral{
		Token:    tok,
		Size:     len(contents),
		Type:     collectionType,
		Elements: contents,
	}
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	expression := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	expression.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return expression
}
