package parser

import (
	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/token"
)

func (p *Parser) parseStructStatement() ast.Statement {
	str := &ast.StructStatement{
		Token: p.curToken,
	}

	p.nextToken()

	if !p.curTokenIs(token.IDENT) {
		return nil
	}

	str.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	str.Fields = []*ast.StructField{}
	str.TypeFunctions = []*ast.FunctionStatement{}

	p.nextToken()
	for !p.curTokenIs(token.RBRACE) {
		if p.curTokenIs(token.EOF) {
			return nil
		}

		switch p.curToken.Type {
		case token.PUB, token.IDENT:
			res := p.parseStructField()
			field, ok := res.(*ast.StructField)
			if !ok {
				return nil
			}

			str.Fields = append(str.Fields, field)
		case token.FN:
			res := p.parseFunctionStatement()
			fn, ok := res.(*ast.FunctionStatement)
			if !ok {
				return nil
			}

			str.TypeFunctions = append(str.TypeFunctions, fn)
		case token.IMPL:
			if str.Impl != nil {
				// Double impl block error
				return nil
			}
			res := p.parseStructImpl()

			impl, ok := res.(*ast.StructImplBlock)
			if !ok {
				return nil
			}

			str.Impl = impl
		default:
			return nil
		}

		p.nextToken()
	}

	return str
}

func (p *Parser) parseStructField() ast.Node {
	field := &ast.StructField{}

	if p.curTokenIs(token.PUB) {
		field.Public = true
		p.nextToken()
	}

	field.Token = p.curToken
	field.Name = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	p.nextToken()

	if p.curTokenIs(token.COLON) {
		declaredType, ok := p.parseTypeAfterColon()
		if !ok {
			return nil
		}

		field.Type = declaredType
		p.nextToken()
	}

	if p.curTokenIs(token.ASSIGN) {
		p.nextToken()
		field.Value = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return field
}

func (p *Parser) parseStructImpl() ast.Statement {
	impl := &ast.StructImplBlock{
		Token:   p.curToken,
		Methods: []*ast.FunctionStatement{},
	}

	p.nextToken()

	if !p.curTokenIs(token.LBRACE) {
		return nil
	}

	p.nextToken()
	for !p.curTokenIs(token.RBRACE) {
		if p.curTokenIs(token.EOF) {
			return nil
		}

		switch p.curToken.Type {
		case token.FN:
			res := p.parseFunctionStatement()
			fn, ok := res.(*ast.FunctionStatement)
			if !ok {
				return nil
			}

			impl.Methods = append(impl.Methods, fn)
		default:
			return nil
		}

		p.nextToken()
	}

	return impl
}

func (p *Parser) parseStructLiteral(str ast.Expression) ast.Expression {
	name, ok := str.(*ast.Identifier)
	if !ok {
		p.errors = append(p.errors, "struct literal type must be an identifier")
		return nil
	}

	strLit := &ast.StructLiteral{
		Token: name.Token,
		Name:  name,
	}
	strLit.Fields = p.parseStructLiteralArguments()
	return strLit
}

func (p *Parser) parseStructLiteralArguments() []*ast.StructLiteralField {
	args := []*ast.StructLiteralField{}

	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return args
	}

	for {
		p.nextToken()
		if !p.curTokenIs(token.IDENT) {
			p.errors = append(p.errors, "expected struct literal field name")
			return nil
		}

		field := &ast.StructLiteralField{
			Token: p.curToken,
			Name: &ast.Identifier{
				Token: p.curToken,
				Value: p.curToken.Literal,
			},
		}

		if p.peekTokenIs(token.COLON) {
			p.nextToken()
			p.nextToken()
			field.Value = p.parseExpression(LOWEST)
			if field.Value == nil {
				return nil
			}
		} else {
			field.Shorthand = true
			field.Value = field.Name
		}

		args = append(args, field)

		if p.peekTokenIs(token.RBRACE) {
			p.nextToken()
			return args
		}

		if !p.expectPeek(token.COMMA) {
			return nil
		}

		if p.peekTokenIs(token.RBRACE) {
			p.nextToken()
			return args
		}
	}
}
