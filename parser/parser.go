package parser

import (
	"fmt"
	"luminalang/ast"
	"luminalang/token"
	"strconv"
)

type Parser struct {
	tokens []token.Token
	pos    int
}

func New(tokens []token.Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

func (p *Parser) Parse() (*ast.Program, error) {
	prog := &ast.Program{}
	for !p.atEnd() {
		if p.check(token.TK_USE) {
			useDecl, err := p.parseUseDecl()
			if err != nil {
				return nil, err
			}
			prog.Imports = append(prog.Imports, *useDecl)
		} else if p.check(token.TK_FN) {
			fn, err := p.parseFnDecl()
			if err != nil {
				return nil, err
			}
			prog.Functions = append(prog.Functions, *fn)
		} else {
			return nil, fmt.Errorf("%d:%d: syntax error, expected 'fn' or 'use'", p.peek().Line, p.peek().Column)
		}
	}
	return prog, nil
}

func (p *Parser) parseUseDecl() (*ast.UseDecl, error) {
	p.advance()
	modTok, err := p.expect(token.TK_IDENT, "expected module name after 'use'")
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.TK_DCOLON, "expected '::' after module name"); err != nil {
		return nil, err
	}
	if p.check(token.TK_STAR) {
		p.advance()
		return &ast.UseDecl{
			Module:   modTok.Lexeme,
			Wildcard: true,
		}, nil
	}
	var items []ast.ImportItem
	for {
		itemTok, err := p.expect(token.TK_IDENT, "expected item name after '::'")
		if err != nil {
			return nil, err
		}
		item := ast.ImportItem{Name: itemTok.Lexeme}
		if p.match(token.TK_AS) {
			aliasTok, err := p.expect(token.TK_IDENT, "expected alias name after 'as'")
			if err != nil {
				return nil, err
			}
			item.Alias = aliasTok.Lexeme
		}
		items = append(items, item)
		if !p.match(token.TK_COMMA) {
			break
		}
	}
	return &ast.UseDecl{
		Module:   modTok.Lexeme,
		Items:    items,
		Wildcard: false,
	}, nil
}

func (p *Parser) peek() token.Token {
	return p.tokens[p.pos]
}

func (p *Parser) previous() token.Token {
	return p.tokens[p.pos-1]
}

func (p *Parser) advance() token.Token {
	tok := p.tokens[p.pos]
	if !p.atEnd() {
		p.pos++
	}
	return tok
}

func (p *Parser) atEnd() bool {
	return p.peek().Type == token.TK_EOF
}

func (p *Parser) check(t token.TokenType) bool {
	if p.atEnd() {
		return false
	}
	return p.peek().Type == t
}

func (p *Parser) match(types ...token.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) expect(t token.TokenType, msg string) (token.Token, error) {
	if p.check(t) {
		return p.advance(), nil
	}
	return token.Token{}, fmt.Errorf("%d:%d: syntax error, %s", p.peek().Line, p.peek().Column, msg)
}

func (p *Parser) parseTypeName() (token.Token, error) {
	if p.match(token.TK_LBRACKET) {
		if _, err := p.expect(token.TK_RBRACKET, "expected ']'"); err != nil {
			return token.Token{}, err
		}
		if !p.check(token.TK_I32) && !p.check(token.TK_F64) && !p.check(token.TK_BOOL) && !p.check(token.TK_STRING_TYPE) {
			return token.Token{}, fmt.Errorf("%d:%d: syntax error, expected element type 'i32', 'f64', 'bool' or 'string' after '[]'", p.peek().Line, p.peek().Column)
		}
		t := p.advance()
		t.Lexeme = "[]" + t.Lexeme
		return t, nil
	}
	if !p.check(token.TK_I32) && !p.check(token.TK_F64) && !p.check(token.TK_BOOL) && !p.check(token.TK_STRING_TYPE) {
		return token.Token{}, fmt.Errorf("%d:%d: syntax error, expected type 'i32', 'f64', 'bool' or 'string'", p.peek().Line, p.peek().Column)
	}
	return p.advance(), nil
}

func (p *Parser) parseFnDecl() (*ast.FnDecl, error) {
	start := p.peek()
	if _, err := p.expect(token.TK_FN, "expected 'fn'"); err != nil {
		return nil, err
	}
	nameTok, err := p.expect(token.TK_IDENT, "expected function name")
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.TK_LPAREN, "expected '('"); err != nil {
		return nil, err
	}
	params, err := p.parseParams()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.TK_RPAREN, "expected ')'"); err != nil {
		return nil, err
	}
	p.match(token.TK_COLON)
	retTok, err := p.parseTypeName()
	if err != nil {
		return nil, err
	}
	retType := retTok.Lexeme
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.FnDecl{
		PosInfo: ast.PosInfo{Line: start.Line, Column: start.Column},
		Name:    nameTok.Lexeme,
		Params:  params,
		RetType: retType,
		Body:    body,
	}, nil
}

func (p *Parser) parseParams() ([]ast.Param, error) {
	var params []ast.Param
	if p.check(token.TK_RPAREN) {
		return params, nil
	}
	for {
		nameTok, err := p.expect(token.TK_IDENT, "expected parameter name")
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(token.TK_COLON, "expected ':'"); err != nil {
			return nil, err
		}
		typeTok, err := p.parseTypeName()
		if err != nil {
			return nil, err
		}
		params = append(params, ast.Param{Name: nameTok.Lexeme, Type: typeTok.Lexeme})
		if !p.match(token.TK_COMMA) {
			break
		}
	}
	return params, nil
}

func (p *Parser) parseBlock() ([]ast.Stmt, error) {
	if _, err := p.expect(token.TK_LBRACE, "expected '{'"); err != nil {
		return nil, err
	}
	var stmts []ast.Stmt
	for !p.check(token.TK_RBRACE) && !p.atEnd() {
		s, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, s)
	}
	if _, err := p.expect(token.TK_RBRACE, "expected '}'"); err != nil {
		return nil, err
	}
	return stmts, nil
}

func (p *Parser) parseStmt() (ast.Stmt, error) {
	switch {
	case p.check(token.TK_LET):
		return p.parseLet()
	case p.check(token.TK_IF):
		return p.parseIf()
	case p.check(token.TK_WHILE):
		return p.parseWhile()
	case p.check(token.TK_OUT):
		return p.parseOut()
	case p.check(token.TK_BACK):
		return p.parseBack()
	case p.check(token.TK_HALT):
		tok := p.peek()
		p.advance()
		return &ast.HaltStmt{PosInfo: ast.PosInfo{Line: tok.Line, Column: tok.Column}}, nil
	case p.check(token.TK_SKIP):
		tok := p.peek()
		p.advance()
		return &ast.SkipStmt{PosInfo: ast.PosInfo{Line: tok.Line, Column: tok.Column}}, nil
	case p.check(token.TK_FOR):
		return p.parseForIn()
	case p.check(token.TK_IDENT):
		return p.parseAssignOrCall()
	default:
		return nil, fmt.Errorf("%d:%d: syntax error, unexpected token '%s'", p.peek().Line, p.peek().Column, p.peek().Lexeme)
	}
}

func (p *Parser) parseLet() (ast.Stmt, error) {
	start := p.peek()
	p.advance()
	nameTok, err := p.expect(token.TK_IDENT, "expected variable name")
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.TK_COLON, "expected ':'"); err != nil {
		return nil, err
	}
	typeTok, err := p.parseTypeName()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.TK_ASSIGN, "expected '='"); err != nil {
		return nil, err
	}
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &ast.LetStmt{PosInfo: ast.PosInfo{Line: start.Line, Column: start.Column}, Name: nameTok.Lexeme, Type: typeTok.Lexeme, Init: expr}, nil
}

func (p *Parser) parseAssignOrCall() (ast.Stmt, error) {
	nameTok := p.advance()
	pos := ast.PosInfo{Line: nameTok.Line, Column: nameTok.Column}
	if p.check(token.TK_LPAREN) {
		p.pos--
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &ast.ExprStmt{PosInfo: pos, Expr: expr}, nil
	}
	if p.check(token.TK_ASSIGN) {
		p.advance()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &ast.AssignStmt{PosInfo: pos, Name: nameTok.Lexeme, Expr: expr}, nil
	}
	p.pos--
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.match(token.TK_ASSIGN) {
		idx, ok := expr.(*ast.IndexExpr)
		if !ok {
			return nil, fmt.Errorf("%d:%d: syntax error, invalid assignment target", nameTok.Line, nameTok.Column)
		}
		tgt, ok := idx.Target.(*ast.Ident)
		if !ok {
			return nil, fmt.Errorf("%d:%d: syntax error, invalid assignment target", nameTok.Line, nameTok.Column)
		}
		val, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &ast.IndexAssignStmt{PosInfo: pos, Name: tgt.Name, Index: idx.Index, Value: val}, nil
	}
	return &ast.ExprStmt{PosInfo: pos, Expr: expr}, nil
}

func (p *Parser) parseIf() (ast.Stmt, error) {
	start := p.peek()
	p.advance()
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	thenBody, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	var elifConds []ast.Expr
	var elifBodies [][]ast.Stmt
	for p.check(token.TK_ELIF) {
		p.advance()
		econd, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		ebody, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		elifConds = append(elifConds, econd)
		elifBodies = append(elifBodies, ebody)
	}
	var elseBody []ast.Stmt
	if p.match(token.TK_ELSE) {
		elseBody, err = p.parseBlock()
		if err != nil {
			return nil, err
		}
	}
	return &ast.IfStmt{
		PosInfo:    ast.PosInfo{Line: start.Line, Column: start.Column},
		Cond:       cond,
		Then:       thenBody,
		ElifConds:  elifConds,
		ElifBodies: elifBodies,
		ElseBody:   elseBody,
	}, nil
}

func (p *Parser) parseWhile() (ast.Stmt, error) {
	start := p.peek()
	p.advance()
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.WhileStmt{PosInfo: ast.PosInfo{Line: start.Line, Column: start.Column}, Cond: cond, Body: body}, nil
}

func (p *Parser) parseOut() (ast.Stmt, error) {
	start := p.peek()
	p.advance()
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &ast.OutStmt{PosInfo: ast.PosInfo{Line: start.Line, Column: start.Column}, Expr: expr}, nil
}

func (p *Parser) parseBack() (ast.Stmt, error) {
	start := p.peek()
	p.advance()
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &ast.BackStmt{PosInfo: ast.PosInfo{Line: start.Line, Column: start.Column}, Expr: expr}, nil
}

func (p *Parser) parseExpr() (ast.Expr, error) {
	return p.logicOr()
}

func (p *Parser) logicOr() (ast.Expr, error) {
	expr, err := p.logicAnd()
	if err != nil {
		return nil, err
	}
	for p.match(token.TK_OR) {
		right, err := p.logicAnd()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpr{Left: expr, Op: token.TK_OR, Right: right}
	}
	return expr, nil
}

func (p *Parser) logicAnd() (ast.Expr, error) {
	expr, err := p.bitOr()
	if err != nil {
		return nil, err
	}
	for p.match(token.TK_AND) {
		right, err := p.bitOr()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpr{Left: expr, Op: token.TK_AND, Right: right}
	}
	return expr, nil
}

func (p *Parser) bitOr() (ast.Expr, error) {
	expr, err := p.bitXor()
	if err != nil {
		return nil, err
	}
	for p.match(token.TK_PIPE) {
		right, err := p.bitXor()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpr{Left: expr, Op: token.TK_PIPE, Right: right}
	}
	return expr, nil
}

func (p *Parser) bitXor() (ast.Expr, error) {
	expr, err := p.bitAnd()
	if err != nil {
		return nil, err
	}
	for p.match(token.TK_CARET) {
		right, err := p.bitAnd()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpr{Left: expr, Op: token.TK_CARET, Right: right}
	}
	return expr, nil
}

func (p *Parser) bitAnd() (ast.Expr, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}
	for p.match(token.TK_AMP) {
		right, err := p.equality()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpr{Left: expr, Op: token.TK_AMP, Right: right}
	}
	return expr, nil
}

func (p *Parser) equality() (ast.Expr, error) {
	expr, err := p.comparison()
	if err != nil {
		return nil, err
	}
	for p.match(token.TK_EQ, token.TK_NEQ) {
		op := p.previous().Type
		right, err := p.comparison()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpr{Left: expr, Op: op, Right: right}
	}
	return expr, nil
}

func (p *Parser) comparison() (ast.Expr, error) {
	expr, err := p.shift()
	if err != nil {
		return nil, err
	}
	for p.match(token.TK_LT, token.TK_GT, token.TK_LE, token.TK_GE) {
		op := p.previous().Type
		right, err := p.shift()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpr{Left: expr, Op: op, Right: right}
	}
	return expr, nil
}

func (p *Parser) shift() (ast.Expr, error) {
	expr, err := p.addition()
	if err != nil {
		return nil, err
	}
	for p.match(token.TK_SHL, token.TK_SHR) {
		op := p.previous().Type
		right, err := p.addition()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpr{Left: expr, Op: op, Right: right}
	}
	return expr, nil
}

func (p *Parser) addition() (ast.Expr, error) {
	expr, err := p.multiplication()
	if err != nil {
		return nil, err
	}
	for p.match(token.TK_PLUS, token.TK_MINUS) {
		op := p.previous().Type
		right, err := p.multiplication()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpr{Left: expr, Op: op, Right: right}
	}
	return expr, nil
}

func (p *Parser) multiplication() (ast.Expr, error) {
	expr, err := p.unary()
	if err != nil {
		return nil, err
	}
	for p.match(token.TK_STAR, token.TK_SLASH, token.TK_PERCENT) {
		op := p.previous().Type
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpr{Left: expr, Op: op, Right: right}
	}
	return expr, nil
}

func (p *Parser) unary() (ast.Expr, error) {
	if p.match(token.TK_MINUS, token.TK_TILDE, token.TK_NOT) {
		op := p.previous().Type
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{Op: op, Right: right}, nil
	}
	return p.indexed()
}

func (p *Parser) indexed() (ast.Expr, error) {
	expr, err := p.primary()
	if err != nil {
		return nil, err
	}
	for p.match(token.TK_LBRACKET) {
		start := p.previous()
		idx, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(token.TK_RBRACKET, "expected ']'"); err != nil {
			return nil, err
		}
		expr = &ast.IndexExpr{PosInfo: ast.PosInfo{Line: start.Line, Column: start.Column}, Target: expr, Index: idx}
	}
	return expr, nil
}

func (p *Parser) primary() (ast.Expr, error) {
	if p.match(token.TK_LPAREN) {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(token.TK_RPAREN, "expected ')'"); err != nil {
			return nil, err
		}
		return expr, nil
	}
	if p.match(token.TK_LBRACKET) {
		start := p.previous()
		var elems []ast.Expr
		if !p.check(token.TK_RBRACKET) {
			for {
				e, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				elems = append(elems, e)
				if !p.match(token.TK_COMMA) {
					break
				}
			}
		}
		if _, err := p.expect(token.TK_RBRACKET, "expected ']'"); err != nil {
			return nil, err
		}
		return &ast.SliceLiteral{PosInfo: ast.PosInfo{Line: start.Line, Column: start.Column}, Elems: elems}, nil
	}
	if p.match(token.TK_INT) {
		return &ast.IntLiteral{Value: p.previous().Value}, nil
	}
	if p.match(token.TK_FLOAT) {
		f, err := strconv.ParseFloat(p.previous().Lexeme, 64)
		if err != nil {
			return nil, fmt.Errorf("%d:%d: invalid float literal '%s'", p.previous().Line, p.previous().Column, p.previous().Lexeme)
		}
		return &ast.FloatLiteral{Value: f}, nil
	}
	if p.match(token.TK_STRING) {
		return &ast.StringLiteral{Value: p.previous().Lexeme}, nil
	}
	if p.match(token.TK_TRUE) {
		return &ast.BoolLiteral{Value: true}, nil
	}
	if p.match(token.TK_FALSE) {
		return &ast.BoolLiteral{Value: false}, nil
	}
	if p.match(token.TK_INPUT) {
		if _, err := p.expect(token.TK_LPAREN, "expected '('"); err != nil {
			return nil, err
		}
		if _, err := p.expect(token.TK_RPAREN, "expected ')'"); err != nil {
			return nil, err
		}
		return &ast.InputExpr{}, nil
	}
	if p.match(token.TK_IDENT) {
		name := p.previous().Lexeme
		if p.match(token.TK_LPAREN) {
			var args []ast.Expr
			if !p.check(token.TK_RPAREN) {
				for {
					arg, err := p.parseExpr()
					if err != nil {
						return nil, err
					}
					args = append(args, arg)
					if !p.match(token.TK_COMMA) {
						break
					}
				}
			}
			if _, err := p.expect(token.TK_RPAREN, "expected ')'"); err != nil {
				return nil, err
			}
			return &ast.CallExpr{Callee: name, Args: args}, nil
		}
		return &ast.Ident{Name: name}, nil
	}
	return nil, fmt.Errorf("%d:%d: syntax error, expected expression", p.peek().Line, p.peek().Column)
}

func (p *Parser) parseForIn() (ast.Stmt, error) {
	start := p.peek()
	p.advance()
	varName, err := p.expect(token.TK_IDENT, "expected variable name after 'for'")
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.TK_IN, "expected 'in' after variable name"); err != nil {
		return nil, err
	}
	iterExpr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.ForInStmt{
		PosInfo:  ast.PosInfo{Line: start.Line, Column: start.Column},
		VarName:  varName.Lexeme,
		IterExpr: iterExpr,
		Body:     body,
	}, nil
}
