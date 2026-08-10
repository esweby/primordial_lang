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
			p.addDiagnostic("P1501", "invalid array length "+p.curToken.Literal, p.curToken)
			return nil
		}

		length = int(parsedInt)
		if length <= 0 {
			p.addDiagnostic("P1502", "array length must be greater than zero", p.curToken)
			return nil
		}

		p.nextToken()
	}

	if !p.curTokenIs(token.RBRACKET) {
		p.addDiagnostic("P1503", "expected array length or ']', found "+describeToken(p.curToken), p.curToken, token.INT_LITERAL, token.RBRACKET)
		return nil
	}

	p.nextToken()
	collectionType, ok := types.GetBuiltin(p.curToken.Literal)
	if !ok {
		p.addDiagnostic("P1504", "expected collection element type, found "+describeToken(p.curToken), p.curToken, token.IDENT)
		return nil
	}

	p.nextToken()
	if !p.curTokenIs(token.LBRACE) {
		p.addDiagnostic("P1505", "expected '{' after collection type, found "+describeToken(p.curToken), p.curToken, token.LBRACE)
		return nil
	}

	p.nextToken()
	contents := []ast.Expression{}
	if p.curTokenIs(token.RBRACE) {
		return p.returnCollection(tok, length, collectionType, contents)
	}

	for {
		before := len(p.diagnostics)
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			p.ensureDiagnostic(before, "P1506", "expected collection element", p.curToken)
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
	before := len(p.diagnostics)
	expression.Index = p.parseExpression(LOWEST)
	if expression.Index == nil {
		p.ensureDiagnostic(before, "P1507", "expected index expression", p.curToken)
		return nil
	}

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return expression
}
