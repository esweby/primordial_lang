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
		p.addDiagnostic("P1301", "expected struct name, found "+describeToken(p.curToken), p.curToken, token.IDENT)
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
			p.addDiagnostic("P1302", "expected '}' to close struct declaration", p.curToken, token.RBRACE)
			return nil
		}

		switch p.curToken.Type {
		case token.PUB, token.IDENT:
			before := len(p.diagnostics)
			res := p.parseStructField()
			field, ok := res.(*ast.StructField)
			if !ok {
				p.ensureDiagnostic(before, "P1303", "invalid struct field declaration", p.curToken)
				return nil
			}

			str.Fields = append(str.Fields, field)
		case token.MUT:
			p.addDiagnostic("P1304", "struct fields only support the pub modifier", p.curToken, token.PUB, token.IDENT)
			return nil
		case token.FN:
			before := len(p.diagnostics)
			res := p.parseFunctionStatement()
			fn, ok := res.(*ast.FunctionStatement)
			if !ok {
				p.ensureDiagnostic(before, "P1305", "invalid struct type function", p.curToken)
				return nil
			}

			str.TypeFunctions = append(str.TypeFunctions, fn)
		case token.IMPL:
			if str.Impl != nil {
				p.addDiagnostic("P1306", "struct may only contain one impl block", p.curToken)
				return nil
			}
			before := len(p.diagnostics)
			res := p.parseStructImpl()

			impl, ok := res.(*ast.StructImplBlock)
			if !ok {
				p.ensureDiagnostic(before, "P1307", "invalid struct impl block", p.curToken)
				return nil
			}

			str.Impl = impl
		default:
			p.addDiagnostic("P1308", "expected struct field, function, impl block, or '}', found "+describeToken(p.curToken), p.curToken, token.IDENT, token.FN, token.IMPL, token.RBRACE)
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

	if !p.curTokenIs(token.IDENT) {
		p.addDiagnostic("P1309", "expected struct field name, found "+describeToken(p.curToken), p.curToken, token.IDENT)
		return nil
	}

	field.Token = p.curToken
	field.Name = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	p.nextToken()

	if p.curTokenIs(token.COLON) {
		before := len(p.diagnostics)
		declaredType, ok := p.parseTypeAfterColon()
		if !ok {
			p.ensureDiagnostic(before, "P1310", "expected struct field type", p.curToken, token.IDENT)
			return nil
		}

		field.Type = declaredType
		p.nextToken()
	}

	if p.curTokenIs(token.ASSIGN) {
		p.nextToken()
		before := len(p.diagnostics)
		field.Value = p.parseExpression(LOWEST)
		if field.Value == nil {
			p.ensureDiagnostic(before, "P1311", "expected default value for struct field", p.curToken)
			return nil
		}
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
		p.addDiagnostic("P1312", "expected '{' after 'impl', found "+describeToken(p.curToken), p.curToken, token.LBRACE)
		return nil
	}

	p.nextToken()
	for !p.curTokenIs(token.RBRACE) {
		if p.curTokenIs(token.EOF) {
			p.addDiagnostic("P1313", "expected '}' to close impl block", p.curToken, token.RBRACE)
			return nil
		}

		switch p.curToken.Type {
		case token.FN:
			before := len(p.diagnostics)
			res := p.parseFunctionStatement()
			fn, ok := res.(*ast.FunctionStatement)
			if !ok {
				p.ensureDiagnostic(before, "P1314", "invalid method declaration", p.curToken)
				return nil
			}

			impl.Methods = append(impl.Methods, fn)
		default:
			p.addDiagnostic("P1315", "expected method declaration or '}', found "+describeToken(p.curToken), p.curToken, token.FN, token.RBRACE)
			return nil
		}

		p.nextToken()
	}

	return impl
}

func (p *Parser) parseStructLiteral(str ast.Expression) ast.Expression {
	name, ok := str.(*ast.Identifier)
	if !ok {
		p.addDiagnostic("P1320", "struct literal type must be an identifier", p.curToken, token.IDENT)
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
			p.addDiagnostic("P1321", "expected struct literal field name, found "+describeToken(p.curToken), p.curToken, token.IDENT)
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
			before := len(p.diagnostics)
			field.Value = p.parseExpression(LOWEST)
			if field.Value == nil {
				p.ensureDiagnostic(before, "P1322", "expected value for struct literal field "+field.Name.Value, p.curToken)
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
