package parser

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/token"
)

func (p *Parser) parseLabelledForStatement() ast.Statement {
	fmt.Printf("DEBUG: parseLabelledForStatement called, curToken=%s peekToken=%s\n",
		p.curToken.Literal, p.peekToken.Literal)
	exprStmt := &ast.ExpressionStatement{Token: p.curToken}

	forLoop := p.parseForExpression()
	if forLoop == nil {
		fmt.Println("DEBUG: parseForExpression returned nil")
	}
	exprStmt.Expression = forLoop
	return exprStmt
}

func (p *Parser) parseForExpression() ast.Expression {
	forLoop := &ast.ForLoop{
		Token: p.curToken,
	}

	if p.curTokenIs(token.IDENT) {
		if !p.peekTokenIs(token.COLON) {
			return nil
		}

		forLoop.Label = &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}

		p.nextToken() // should be colon
		p.nextToken() // should be for
	}

	if !p.curTokenIs(token.FOR) {
		fmt.Println("DEBUG: parseForExpression returning nil at token.FOR")
		fmt.Printf("DEBUG: parseForExpression, curToken=%s peekToken=%s\n",
			p.curToken.Literal, p.peekToken.Literal)
		// expecting for token error
		return nil
	}

	p.nextToken()

	ctrl := p.parseForController()
	// error check
	if ctrl == nil {
		return nil
	}

	forLoop.Controller = ctrl

	forLoop.Body = p.parseBlockExpression()

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return forLoop
}

func (p *Parser) parseForController() ast.ForController {
	if p.curTokenIs(token.LBRACE) {
		return &ast.Infinite{Token: p.curToken}
	}

	if p.curTokenIs(token.IDENT) {
		// range loop (unchanged, but ensure it works)
		rnge := &ast.Range{
			Token:     p.curToken,
			Variables: []*ast.Identifier{},
		}

		rnge.Variables = append(rnge.Variables, &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		})

		p.nextToken()

		if p.curTokenIs(token.COMMA) {
			p.nextToken()
			if !p.curTokenIs(token.IDENT) {
				return nil
			}
			rnge.Variables = append(rnge.Variables, &ast.Identifier{
				Token: p.curToken,
				Value: p.curToken.Literal,
			})
			p.nextToken()
		}

		if !p.curTokenIs(token.DECLARE) {
			return nil
		}
		p.nextToken()

		if !p.curTokenIs(token.RANGE) {
			return nil
		}
		p.nextToken()

		p.disableStructLiteralParsing()
		iterable := p.parseExpression(LOWEST)
		if iterable == nil {
			return nil
		}
		p.enableStructLiteralParsing()

		// After parseExpression, curToken is last token of iterable.
		// We need to advance to the body's '{'
		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		rnge.Iterable = iterable
		return rnge
	}

	if p.curTokenIs(token.LPAREN) {
		openParen := p.curToken
		p.nextToken() // move past '(' to first token inside

		// Check for constructed loop: identifier followed by ':=' (declaration)
		if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.DECLARE) {
			// --- constructed loop ---
			// Parse initializer manually: ident := expr
			initName := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
			p.nextToken() // move to ':='

			if !p.curTokenIs(token.DECLARE) {
				return nil
			}
			initToken := p.curToken
			p.nextToken() // move to init expression start

			initValue := p.parseExpression(LOWEST)
			if initValue == nil {
				return nil
			}

			// Build DeclareStatement for initializer
			initStmt := &ast.DeclareStatement{
				Token: initToken,
				Name:  initName,
				Value: initValue,
				// Type and mutability can be omitted for now
			}

			// After init expression, expect semicolon
			if !p.expectPeek(token.SEMICOLON) {
				return nil
			}
			// Now curToken is ';', advance to condition start
			p.nextToken()

			// Parse condition expression
			cond := p.parseExpression(LOWEST)
			if cond == nil {
				return nil
			}

			// After condition, expect semicolon
			if !p.expectPeek(token.SEMICOLON) {
				return nil
			}
			// Advance to iterator start
			p.nextToken()

			// Parse iterator statement (could be assignment, etc.)
			// For now, we parse as expression statement (or assignment)
			iterStmt := p.parseStatement()
			if iterStmt == nil {
				return nil
			}

			// After iterator, expect ')'
			if !p.expectPeek(token.RPAREN) {
				return nil
			}
			// Expect '{' for body
			if !p.expectPeek(token.LBRACE) {
				return nil
			}

			return &ast.Constructed{
				Token:       openParen,
				Initializer: initStmt,
				Condition:   cond,
				Iterator:    iterStmt,
			}
		}

		// --- while loop ---
		cond := p.parseExpression(LOWEST)
		if cond == nil {
			return nil
		}

		// After condition, expect ')'
		if !p.expectPeek(token.RPAREN) {
			return nil
		}
		// Expect '{' for body
		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		return &ast.While{
			Token:     openParen,
			Condition: cond,
		}
	}

	return nil
}

func (p *Parser) parseContinueStatement() ast.Statement {
	continueStmt := &ast.ContinueStatement{
		Token: p.curToken,
	}

	p.nextToken()

	if !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
		if p.curTokenIs(token.EOF) {
			return continueStmt
		}

	}

	return continueStmt
}

func (p *Parser) parseBreakStatement() ast.Statement {
	breakStmt := &ast.BreakStatement{
		Token: p.curToken,
	}

	p.nextToken()

	if p.curTokenIs(token.IDENT) {
		breakStmt.Label = &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}
		p.nextToken()
	}

	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		expr := p.parseExpression(LOWEST)

		if expr == nil {
			return nil
		}

		if !p.expectPeek(token.RPAREN) {
			return nil
		}

		if !p.expectPeek(token.SEMICOLON) {
			return nil
		}

		breakStmt.Value = expr
	}

	return breakStmt
}
