package compiler

import (
	"fmt"
	"luminalang/ast"
	"luminalang/token"
)

func (c *Compiler) collectLetsInStmt(ctx *fnCtx, st ast.Stmt) error {
	switch v := st.(type) {
	case *ast.LetStmt:
		t, err := c.inferExprType(ctx, v.Init)
		if err != nil {
			return err
		}
		if old, ok := ctx.varTypes[v.Name]; ok && old != t {
			return fmt.Errorf("compile error: variable '%s' redeclared as %s, previously %s", v.Name, t, old)
		}
		ctx.varTypes[v.Name] = t
		return nil
	case *ast.IfStmt:
		for _, b := range v.Then {
			if err := c.collectLetsInStmt(ctx, b); err != nil {
				return err
			}
		}
		for _, b := range v.ElifBodies {
			for _, s := range b {
				if err := c.collectLetsInStmt(ctx, s); err != nil {
					return err
				}
			}
		}
		for _, s := range v.ElseBody {
			if err := c.collectLetsInStmt(ctx, s); err != nil {
				return err
			}
		}
	case *ast.WhileStmt:
		for _, s := range v.Body {
			if err := c.collectLetsInStmt(ctx, s); err != nil {
				return err
			}
		}
	case *ast.ForInStmt:
		ctx.forVars[v.VarName] = "loopvar"
		for _, s := range v.Body {
			if err := c.collectLetsInStmt(ctx, s); err != nil {
				return err
			}
		}
		delete(ctx.forVars, v.VarName)
	}
	return nil
}

func (c *Compiler) inferExprType(ctx *fnCtx, e ast.Expr) (Type, error) {
	switch v := e.(type) {
	case *ast.IntLiteral:
		return TInt, nil
	case *ast.StringLiteral:
		return TStr, nil
	case *ast.BoolLiteral:
		return TBool, nil
	case *ast.Ident:
		if _, ok := ctx.forVars[v.Name]; ok {
			return TStr, nil
		}
		t, ok := ctx.varTypes[v.Name]
		if !ok {
			return TInvalid, fmt.Errorf("compile error: undefined variable '%s'", v.Name)
		}
		return t, nil
	case *ast.BinaryExpr:
		lt, err := c.inferExprType(ctx, v.Left)
		if err != nil {
			return TInvalid, err
		}
		rt, err := c.inferExprType(ctx, v.Right)
		if err != nil {
			return TInvalid, err
		}
		switch v.Op {
		case token.TK_PLUS:
			if lt == TStr && rt == TStr {
				return TStr, nil
			}
			if lt == TInt && rt == TInt {
				return TInt, nil
			}
			return TInvalid, fmt.Errorf("compile error: cannot add %s and %s", lt, rt)
		case token.TK_MINUS, token.TK_STAR, token.TK_SLASH:
			if lt != TInt || rt != TInt {
				return TInvalid, fmt.Errorf("compile error: operator requires i32 operands, got %s and %s", lt, rt)
			}
			return TInt, nil
		case token.TK_EQ, token.TK_NEQ:
			if lt != rt {
				return TInvalid, fmt.Errorf("compile error: cannot compare %s and %s", lt, rt)
			}
			return TBool, nil
		case token.TK_LT, token.TK_GT, token.TK_LE, token.TK_GE:
			if lt != TInt || rt != TInt {
				return TInvalid, fmt.Errorf("compile error: relational operator requires i32 operands, got %s and %s", lt, rt)
			}
			return TBool, nil
		}
	case *ast.UnaryExpr:
		rt, err := c.inferExprType(ctx, v.Right)
		if err != nil {
			return TInvalid, err
		}
		if v.Op == token.TK_MINUS && rt != TInt {
			return TInvalid, fmt.Errorf("compile error: cannot negate %s", rt)
		}
		return rt, nil
	case *ast.CallExpr:
		if fn, ok := c.functions[v.Callee]; ok {
			return typeFromName(fn.RetType), nil
		}
		full, ok := c.imports[v.Callee]
		if !ok {
			return TInvalid, fmt.Errorf("compile error: undefined function '%s'", v.Callee)
		}
		return builtins[full].ret, nil
	case *ast.InputExpr:
		return TInt, nil
	}
	return TInvalid, fmt.Errorf("compile error: unknown expression type")
}
