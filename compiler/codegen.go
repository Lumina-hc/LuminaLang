package compiler

import (
	"fmt"
	"luminalang/ast"
	"luminalang/token"
	"strconv"
	"strings"
)

func (c *Compiler) genFunction(fn *ast.FnDecl) (string, error) {
	ctx := &fnCtx{
		c:        c,
		retType:  typeFromName(fn.RetType),
		varTypes: make(map[string]Type),
		forVars:  make(map[string]string),
	}

	var sig strings.Builder
	sig.WriteString("func lumf_")
	sig.WriteString(fn.Name)
	sig.WriteString("(")
	for i, p := range fn.Params {
		if i > 0 {
			sig.WriteString(", ")
		}
		sig.WriteString("lumv_")
		sig.WriteString(p.Name)
		sig.WriteString(" ")
		sig.WriteString(goTypeName(typeFromName(p.Type)))
		ctx.varTypes[p.Name] = typeFromName(p.Type)
	}
	sig.WriteString(") ")
	sig.WriteString(goTypeName(ctx.retType))
	sig.WriteString(" {\n")

	for _, st := range fn.Body {
		if err := c.collectLetsInStmt(ctx, st); err != nil {
			return "", err
		}
	}

	var body strings.Builder
	for _, st := range fn.Body {
		code, err := c.genStmt(ctx, st)
		if err != nil {
			return "", err
		}
		body.WriteString(code)
	}

	var out strings.Builder
	out.WriteString(sig.String())
	for name := range ctx.varTypes {
		if isParamName(fn, name) {
			continue
		}
		out.WriteString("    var lumv_")
		out.WriteString(name)
		out.WriteString(" ")
		out.WriteString(goTypeName(ctx.varTypes[name]))
		out.WriteString("\n")
	}
	out.WriteString(body.String())
	out.WriteString("    return " + defaultReturn(ctx.retType) + "\n")
	out.WriteString("}\n")
	return out.String(), nil
}

func defaultReturn(t Type) string {
	switch t {
	case TBool:
		return "false"
	case TStr:
		return `""`
	default:
		return "0"
	}
}

func isParamName(fn *ast.FnDecl, name string) bool {
	for _, p := range fn.Params {
		if p.Name == name {
			return true
		}
	}
	return false
}

func (c *Compiler) genStmt(ctx *fnCtx, st ast.Stmt) (string, error) {
	switch v := st.(type) {
	case *ast.LetStmt:
		code, t, err := c.genExpr(ctx, v.Init)
		if err != nil {
			return "", err
		}
		declType, ok := ctx.varTypes[v.Name]
		if !ok {
			return "", fmt.Errorf("compile error: unknown variable '%s'", v.Name)
		}
		if t != declType {
			return "", fmt.Errorf("compile error: cannot assign %s to variable '%s' of type %s", t, v.Name, declType)
		}
		return "    lumv_" + v.Name + " = " + code + "\n", nil
	case *ast.AssignStmt:
		code, t, err := c.genExpr(ctx, v.Expr)
		if err != nil {
			return "", err
		}
		vt, ok := ctx.varTypes[v.Name]
		if !ok {
			return "", fmt.Errorf("compile error: undefined variable '%s'", v.Name)
		}
		if vt != t {
			return "", fmt.Errorf("compile error: cannot assign %s to variable '%s' of type %s", t, v.Name, vt)
		}
		return "    lumv_" + v.Name + " = " + code + "\n", nil
	case *ast.OutStmt:
		code, _, err := c.genExpr(ctx, v.Expr)
		if err != nil {
			return "", err
		}
		return "    fmt.Println(" + code + ")\n", nil
	case *ast.BackStmt:
		code, t, err := c.genExpr(ctx, v.Expr)
		if err != nil {
			return "", err
		}
		if ctx.retType != t {
			return "", fmt.Errorf("compile error: function returns %s but 'back' returns %s", ctx.retType, t)
		}
		return "    return " + code + "\n", nil
	case *ast.HaltStmt:
		if ctx.loopDepth > 0 {
			return "    break\n", nil
		}
		return "    os.Exit(0)\n", nil
	case *ast.SkipStmt:
		if ctx.loopDepth > 0 {
			return "    continue\n", nil
		}
		return "", fmt.Errorf("compile error: 'skip' used outside a loop")
	case *ast.IfStmt:
		return c.genIf(ctx, v)
	case *ast.WhileStmt:
		return c.genWhile(ctx, v)
	case *ast.ForInStmt:
		return c.genForIn(ctx, v)
	case *ast.ExprStmt:
		code, _, err := c.genExpr(ctx, v.Expr)
		if err != nil {
			return "", err
		}
		return "    " + code + "\n", nil
	default:
		return "", fmt.Errorf("compile error: unknown statement type")
	}
}

func (c *Compiler) genCond(ctx *fnCtx, e ast.Expr) (string, error) {
	code, t, err := c.genExpr(ctx, e)
	if err != nil {
		return "", err
	}
	switch t {
	case TBool:
		return code, nil
	case TInt:
		return "lum_truthy(" + code + ")", nil
	default:
		return "", fmt.Errorf("compile error: condition must be i32 or bool, got %s", t)
	}
}

func (c *Compiler) genIf(ctx *fnCtx, s *ast.IfStmt) (string, error) {
	var sb strings.Builder
	cond, err := c.genCond(ctx, s.Cond)
	if err != nil {
		return "", err
	}
	sb.WriteString("    if " + cond + " {\n")
	for _, st := range s.Then {
		code, err := c.genStmt(ctx, st)
		if err != nil {
			return "", err
		}
		sb.WriteString(code)
	}
	for i, econd := range s.ElifConds {
		ec, err := c.genCond(ctx, econd)
		if err != nil {
			return "", err
		}
		sb.WriteString("    } else if " + ec + " {\n")
		for _, st := range s.ElifBodies[i] {
			code, err := c.genStmt(ctx, st)
			if err != nil {
				return "", err
			}
			sb.WriteString(code)
		}
	}
	if s.ElseBody != nil {
		sb.WriteString("    } else {\n")
		for _, st := range s.ElseBody {
			code, err := c.genStmt(ctx, st)
			if err != nil {
				return "", err
			}
			sb.WriteString(code)
		}
	}
	sb.WriteString("    }\n")
	return sb.String(), nil
}

func (c *Compiler) genWhile(ctx *fnCtx, s *ast.WhileStmt) (string, error) {
	cond, err := c.genCond(ctx, s.Cond)
	if err != nil {
		return "", err
	}
	sub := *ctx
	sub.loopDepth++
	var sb strings.Builder
	sb.WriteString("    for " + cond + " {\n")
	for _, st := range s.Body {
		code, err := c.genStmt(&sub, st)
		if err != nil {
			return "", err
		}
		sb.WriteString(code)
	}
	sb.WriteString("    }\n")
	return sb.String(), nil
}

func (c *Compiler) genForIn(ctx *fnCtx, s *ast.ForInStmt) (string, error) {
	iter, t, err := c.genExpr(ctx, s.IterExpr)
	if err != nil {
		return "", err
	}
	if t != TStr {
		return "", fmt.Errorf("compile error: 'for...in' requires a string")
	}
	n := c.counter
	c.counter++
	runeName := fmt.Sprintf("lumfr%d", n)
	varName := fmt.Sprintf("lumfv%d", n)

	sub := *ctx
	if sub.forVars == nil {
		sub.forVars = make(map[string]string)
	}
	sub.forVars[s.VarName] = varName
	sub.loopDepth++

	var body strings.Builder
	for _, st := range s.Body {
		code, err := c.genStmt(&sub, st)
		if err != nil {
			return "", err
		}
		body.WriteString(code)
	}

	var sb strings.Builder
	sb.WriteString("    for _, ")
	sb.WriteString(runeName)
	sb.WriteString(" := range ")
	sb.WriteString(iter)
	sb.WriteString(" {\n")
	sb.WriteString("        ")
	sb.WriteString(varName)
	sb.WriteString(" := string(")
	sb.WriteString(runeName)
	sb.WriteString(")\n")
	sb.WriteString(body.String())
	sb.WriteString("    }\n")
	return sb.String(), nil
}

func (c *Compiler) genExpr(ctx *fnCtx, e ast.Expr) (string, Type, error) {
	switch v := e.(type) {
	case *ast.IntLiteral:
		return strconv.FormatInt(v.Value, 10), TInt, nil
	case *ast.StringLiteral:
		return strconv.Quote(v.Value), TStr, nil
	case *ast.BoolLiteral:
		if v.Value {
			return "true", TBool, nil
		}
		return "false", TBool, nil
	case *ast.Ident:
		if gn, ok := ctx.forVars[v.Name]; ok {
			return gn, TStr, nil
		}
		t, ok := ctx.varTypes[v.Name]
		if !ok {
			return "", TInvalid, fmt.Errorf("compile error: undefined variable '%s'", v.Name)
		}
		return "lumv_" + v.Name, t, nil
	case *ast.BinaryExpr:
		return c.genBinary(ctx, v)
	case *ast.UnaryExpr:
		code, t, err := c.genExpr(ctx, v.Right)
		if err != nil {
			return "", TInvalid, err
		}
		if v.Op == token.TK_MINUS {
			if t != TInt {
				return "", TInvalid, fmt.Errorf("compile error: cannot negate %s", t)
			}
			return "(-" + code + ")", TInt, nil
		}
		return code, t, nil
	case *ast.CallExpr:
		return c.genCall(ctx, v)
	case *ast.InputExpr:
		return "lum_input()", TInt, nil
	default:
		return "", TInvalid, fmt.Errorf("compile error: unknown expression type")
	}
}

func (c *Compiler) genBinary(ctx *fnCtx, v *ast.BinaryExpr) (string, Type, error) {
	left, lt, err := c.genExpr(ctx, v.Left)
	if err != nil {
		return "", TInvalid, err
	}
	right, rt, err := c.genExpr(ctx, v.Right)
	if err != nil {
		return "", TInvalid, err
	}
	switch v.Op {
	case token.TK_PLUS:
		if lt == TStr && rt == TStr {
			return "(" + left + " + " + right + ")", TStr, nil
		}
		if lt == TInt && rt == TInt {
			return "(" + left + " + " + right + ")", TInt, nil
		}
		return "", TInvalid, fmt.Errorf("compile error: cannot add %s and %s", lt, rt)
	case token.TK_MINUS, token.TK_STAR:
		if lt != TInt || rt != TInt {
			return "", TInvalid, fmt.Errorf("compile error: operator requires i32 operands, got %s and %s", lt, rt)
		}
		op := "*"
		if v.Op == token.TK_MINUS {
			op = "-"
		}
		return "(" + left + " " + op + " " + right + ")", TInt, nil
	case token.TK_SLASH:
		if lt != TInt || rt != TInt {
			return "", TInvalid, fmt.Errorf("compile error: operator '/' requires i32 operands, got %s and %s", lt, rt)
		}
		return "lum_div(" + left + ", " + right + ")", TInt, nil
	case token.TK_EQ, token.TK_NEQ:
		if lt != rt {
			return "", TInvalid, fmt.Errorf("compile error: cannot compare %s and %s", lt, rt)
		}
		if v.Op == token.TK_EQ {
			return "(" + left + " == " + right + ")", TBool, nil
		}
		return "(" + left + " != " + right + ")", TBool, nil
	case token.TK_LT, token.TK_GT, token.TK_LE, token.TK_GE:
		if lt != TInt || rt != TInt {
			return "", TInvalid, fmt.Errorf("compile error: relational operator requires i32 operands, got %s and %s", lt, rt)
		}
		var op string
		switch v.Op {
		case token.TK_LT:
			op = "<"
		case token.TK_GT:
			op = ">"
		case token.TK_LE:
			op = "<="
		case token.TK_GE:
			op = ">="
		}
		return "(" + left + " " + op + " " + right + ")", TBool, nil
	default:
		return "", TInvalid, fmt.Errorf("compile error: unknown operator")
	}
}

func (c *Compiler) genCall(ctx *fnCtx, v *ast.CallExpr) (string, Type, error) {
	var args []string
	for _, a := range v.Args {
		code, _, err := c.genExpr(ctx, a)
		if err != nil {
			return "", TInvalid, err
		}
		args = append(args, code)
	}
	if fn, ok := c.functions[v.Callee]; ok {
		if len(args) != len(fn.Params) {
			return "", TInvalid, fmt.Errorf("compile error: function '%s' expects %d args, got %d", v.Callee, len(fn.Params), len(args))
		}
		ret := typeFromName(fn.RetType)
		return "lumf_" + v.Callee + "(" + strings.Join(args, ", ") + ")", ret, nil
	}
	full, ok := c.imports[v.Callee]
	if !ok {
		return "", TInvalid, fmt.Errorf("compile error: undefined function '%s'", v.Callee)
	}
	b := builtins[full]
	return "lum_" + strings.ReplaceAll(full, ".", "_") + "(" + strings.Join(args, ", ") + ")", b.ret, nil
}
