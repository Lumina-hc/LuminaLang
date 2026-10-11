package compiler

import (
	"fmt"
	"luminalang/ast"
	"luminalang/token"
)

func (c *Compiler) inferExprType(ctx *fnCtx, e ast.Expr) (Type, error) {
	switch v := e.(type) {
	case *ast.IntLiteral:
		return TInt, nil
	case *ast.FloatLiteral:
		return TFloat, nil
	case *ast.StringLiteral:
		return TStr, nil
	case *ast.BoolLiteral:
		return TBool, nil
	case *ast.Ident:
		s, ok := ctx.lookup(v.Name)
		if !ok {
			return TInvalid, fmt.Errorf("compile error: undefined variable '%s'", v.Name)
		}
		return s.typ, nil
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
			if lt == TFloat && rt == TFloat {
				return TFloat, nil
			}
			return TInvalid, fmt.Errorf("compile error: cannot add %s and %s", lt, rt)
		case token.TK_MINUS, token.TK_STAR, token.TK_SLASH:
			if lt == TInt && rt == TInt {
				return TInt, nil
			}
			if lt == TFloat && rt == TFloat {
				return TFloat, nil
			}
			return TInvalid, fmt.Errorf("compile error: operator requires i32 or f64 operands, got %s and %s", lt, rt)
		case token.TK_PERCENT,
			token.TK_AMP, token.TK_PIPE, token.TK_CARET, token.TK_SHL, token.TK_SHR:
			if lt != TInt || rt != TInt {
				return TInvalid, fmt.Errorf("compile error: operator requires i32 operands, got %s and %s", lt, rt)
			}
			return TInt, nil
		case token.TK_AND, token.TK_OR:
			if lt != TBool || rt != TBool {
				return TInvalid, fmt.Errorf("compile error: logical operator requires bool operands, got %s and %s", lt, rt)
			}
			return TBool, nil
		case token.TK_EQ, token.TK_NEQ:
			if lt != rt {
				return TInvalid, fmt.Errorf("compile error: cannot compare %s and %s", lt, rt)
			}
			return TBool, nil
		case token.TK_LT, token.TK_GT, token.TK_LE, token.TK_GE:
			if (lt == TInt && rt == TInt) || (lt == TFloat && rt == TFloat) {
				return TBool, nil
			}
			return TInvalid, fmt.Errorf("compile error: relational operator requires matching i32 or f64 operands, got %s and %s", lt, rt)
		}
	case *ast.UnaryExpr:
		rt, err := c.inferExprType(ctx, v.Right)
		if err != nil {
			return TInvalid, err
		}
		switch v.Op {
		case token.TK_MINUS:
			if rt != TInt && rt != TFloat {
				return TInvalid, fmt.Errorf("compile error: operator requires i32 or f64 operand, got %s", rt)
			}
			return rt, nil
		case token.TK_TILDE:
			if rt != TInt {
				return TInvalid, fmt.Errorf("compile error: operator requires i32 operand, got %s", rt)
			}
			return TInt, nil
		case token.TK_NOT:
			if rt != TBool {
				return TInvalid, fmt.Errorf("compile error: 'not' requires bool operand, got %s", rt)
			}
			return TBool, nil
		}
		return rt, nil
	case *ast.CallExpr:
		target, fn, ok := c.lookupCall(ctx.module, v.Callee)
		if !ok {
			return c.inferSliceCall(ctx, v)
		}
		if target.kind == 1 {
			if len(v.Args) != len(fn.Params) {
				return TInvalid, fmt.Errorf("compile error: function '%s' expects %d args, got %d", v.Callee, len(fn.Params), len(v.Args))
			}
			for i, a := range v.Args {
				at, err := c.inferExprType(ctx, a)
				if err != nil {
					return TInvalid, err
				}
				pt := typeFromName(fn.Params[i].Type)
				if at != pt {
					return TInvalid, fmt.Errorf("compile error: argument %d of '%s' expects %s, got %s", i+1, v.Callee, pt, at)
				}
			}
			return typeFromName(fn.RetType), nil
		}
		b := builtins[target.builtin]
		if len(v.Args) != len(b.args) {
			return TInvalid, fmt.Errorf("compile error: function '%s' expects %d args, got %d", v.Callee, len(b.args), len(v.Args))
		}
		for i, a := range v.Args {
			at, err := c.inferExprType(ctx, a)
			if err != nil {
				return TInvalid, err
			}
			want := b.args[i]
			if want != TAny && at != want {
				return TInvalid, fmt.Errorf("compile error: argument %d of '%s' expects %s, got %s", i+1, v.Callee, want, at)
			}
		}
		return b.ret, nil
	case *ast.InputExpr:
		return TInt, nil
	case *ast.SliceLiteral:
		if len(v.Elems) == 0 {
			return TInvalid, nil
		}
		elem, err := c.inferExprType(ctx, v.Elems[0])
		if err != nil {
			return TInvalid, err
		}
		return sliceOf(elem), nil
	case *ast.IndexExpr:
		st, err := c.inferExprType(ctx, v.Target)
		if err != nil {
			return TInvalid, err
		}
		elem := sliceElem(st)
		if elem == TInvalid {
			return TInvalid, fmt.Errorf("compile error: cannot index a value of type %s", st)
		}
		it, err := c.inferExprType(ctx, v.Index)
		if err != nil {
			return TInvalid, err
		}
		if it != TInt {
			return TInvalid, fmt.Errorf("compile error: slice index must be i32, got %s", it)
		}
		return elem, nil
	}
	return TInvalid, fmt.Errorf("compile error: unknown expression type")
}

func (c *Compiler) inferSliceCall(ctx *fnCtx, v *ast.CallExpr) (Type, error) {
	switch v.Callee {
	case "len":
		if len(v.Args) != 1 {
			return TInvalid, fmt.Errorf("compile error: 'len' expects 1 argument, got %d", len(v.Args))
		}
		t, err := c.inferExprType(ctx, v.Args[0])
		if err != nil {
			return TInvalid, err
		}
		if t != TStr && !isSlice(t) {
			return TInvalid, fmt.Errorf("compile error: 'len' expects a string or slice, got %s", t)
		}
		return TInt, nil
	case "push", "append":
		if len(v.Args) != 2 {
			return TInvalid, fmt.Errorf("compile error: '%s' expects 2 arguments, got %d", v.Callee, len(v.Args))
		}
		st, err := c.inferExprType(ctx, v.Args[0])
		if err != nil {
			return TInvalid, err
		}
		elem := sliceElem(st)
		if elem == TInvalid {
			return TInvalid, fmt.Errorf("compile error: '%s' expects a slice as first argument, got %s", v.Callee, st)
		}
		vt, err := c.inferExprType(ctx, v.Args[1])
		if err != nil {
			return TInvalid, err
		}
		if vt != elem {
			return TInvalid, fmt.Errorf("compile error: cannot %s %s into %s", v.Callee, vt, st)
		}
		return st, nil
	}
	return TInvalid, fmt.Errorf("compile error: undefined function '%s'", v.Callee)
}
