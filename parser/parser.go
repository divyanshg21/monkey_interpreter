package parser

import (
	"fmt"
	"strconv"

	"github.com/divyanshg21/monkey_interpreter/ast"
	"github.com/divyanshg21/monkey_interpreter/lexer"
	"github.com/divyanshg21/monkey_interpreter/token"
)

const (
	_ int = iota
	LOWEST
	EQUALS       // ==
	LESSGREATER  // > or <
	SUM			 // +
	PRODUCT		 // *
	PREFIX 		 // -X or !X
	CALL		 // myFunction(x)
)

type Parser struct{
	l *lexer.Lexer

	errors []string
	curToken token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns map[token.TokenType]infixParseFn
}

type(
	prefixParseFn func() ast.Expression
	infixParseFn func(ast.Expression) ast.Expression
)

func New(l*lexer.Lexer) *Parser{
	p := &Parser{
		l:	l,
		errors: []string{},
	}
	// p := &Parser{l: l}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)

	p.nextToken()
	p.nextToken()

	return p
}

func(p *Parser) parseIdentifier() ast.Expression{
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}


func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program{
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt !=nil {
			program.Statements = append(program.Statements,stmt)
		}
		p.nextToken()
	}
	return program
}

func (p*Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p*Parser) parseExpressionStatement() *ast.ExpressionStatement{
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression=p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON){
		p.nextToken()
	}

	return stmt
}

func (p*Parser) parseExpression(precedence int) ast.Expression{
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil{
		return nil
	}
	leftExp := prefix()

	return leftExp
}

func (p*Parser) parseLetStatement() *ast.LetStatement{
	stmt:= &ast.LetStatement{Token:p.curToken}

	if !p.expectPeek(token.IDENT){
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN){
		return nil
	}

	for !p.curTokenIs(token.SEMICOLON){
		p.nextToken()
	}
	return stmt
}

func (p*Parser) parseReturnStatement() *ast.ReturnStatement{
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	for !p.curTokenIs(token.SEMICOLON){
		p.nextToken()
	}
	return stmt
}

func (p*Parser) curTokenIs(t token.TokenType) bool{
	return p.curToken.Type == t
}

func (p*Parser) peekTokenIs(t token.TokenType) bool{
	return p.peekToken.Type == t
}

func (p*Parser) expectPeek(t token.TokenType) bool{
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p*Parser) Errors() []string{
	return p.errors
}

func (p*Parser) peekError(t token.TokenType){
	msg := fmt.Sprintf("expected nest token to be as %s, got %s instead", t, p.peekToken.Type)
	p.errors = append (p.errors, msg)
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn){
	p.prefixParseFns[tokenType]=fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn){
	p.infixParseFns[tokenType]=fn
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err!=nil {
		msg := fmt.Sprint("could not parse %q as an integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value

	return lit
}