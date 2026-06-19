package runtime

import (
    "fmt"
    "luminalang/ast"
    "luminalang/token"
    "os"
    "strconv"
    "strings"
)

type Value struct {
    IsString  bool
    IsBool    bool
    IntValue  int64
    StrValue  string
    BoolValue bool
}

func VInt(v int64) Value      { return Value{IntValue: v} }
func VBool(v bool) Value      { return Value{IsBool: true, BoolValue: v} }
func VStr(v string) Value     { return Value{IsString: true, StrValue: v} }

func (v Value) Int() int64   { return v.IntValue }
func (v Value) Bool() bool {
    if v.IsBool {
        return v.BoolValue
    }
    return v.IntValue != 0
}
func (v Value) Str() string   { return v.StrValue }

func ValueString(v Value) string {
    if v.IsString {
        return v.StrValue
    }
    if v.IsBool {
        if v.BoolValue {
            return "true"
        }
        return "false"
    }
    return strconv.FormatInt(v.IntValue, 10)
}

type BuiltinFn func([]Value) (Value, error)

type Environment struct {
    vars   map[string]Value
    parent *Environment
}

func NewEnv(parent *Environment) *Environment {
    return &Environment{vars: make(map[string]Value), parent: parent}
}

func (e *Environment) Get(name string) (Value, bool) {
    if v, ok := e.vars[name]; ok {
        return v, true
    }
    if e.parent != nil {
        return e.parent.Get(name)
    }
    return VInt(0), false
}

func (e *Environment) Set(name string, val Value) {
    e.vars[name] = val
}

type BackSignal struct {
    Value Value
}

func (BackSignal) Error() string { return "__back__" }

type HaltSignal struct{}
type SkipSignal struct{}

func (HaltSignal) Error() string { return "__halt__" }
func (SkipSignal) Error() string { return "__skip__" }

type Interpreter struct {
    functions  map[string]ast.FnDecl
    builtins   map[string]BuiltinFn
    callDepth  int
}

const maxCallDepth = 50000

func NewInterpreter() *Interpreter {
    return &Interpreter{
        functions: make(map[string]ast.FnDecl),
        builtins:  make(map[string]BuiltinFn),
    }
}

func (interp *Interpreter) RegisterBuiltin(name string, fn BuiltinFn) {
    interp.builtins[name] = fn
}

func (interp *Interpreter) ResolveImports(imports []ast.UseDecl) {
    for _, imp := range imports {
        if imp.Wildcard {
            prefix := imp.Module + "."
            for fullName, fn := range interp.builtins {
                if strings.HasPrefix(fullName, prefix) {
                    shortName := strings.TrimPrefix(fullName, prefix)
                    interp.builtins[shortName] = fn
                }
            }
        } else {
            for _, item := range imp.Items {
                fullName := imp.Module + "." + item
                if fn, ok := interp.builtins[fullName]; ok {
                    interp.builtins[item] = fn
                }
            }
        }
    }
}

func (interp *Interpreter) Run(prog *ast.Program) (int, error) {
    for _, fn := range prog.Functions {
        interp.functions[fn.Name] = fn
    }
    mainFn, ok := interp.functions["main"]
    if !ok {
        return 1, fmt.Errorf("main not found")
    }
    result, err := interp.callFn(mainFn.Name, nil)
    if err != nil {
        if _, ok := err.(HaltSignal); ok {
            return 0, nil
        }
        return 1, err
    }
    return int(result.Int()), nil
}

func (interp *Interpreter) callFn(name string, args []Value) (Value, error) {
    if bfn, ok := interp.builtins[name]; ok {
        interp.callDepth++
        defer func() { interp.callDepth-- }()
        return bfn(args)
    }
    fn, ok := interp.functions[name]
    if !ok {
        return VInt(0), fmt.Errorf("type error, undefined function '%s'", name)
    }
    if len(args) != len(fn.Params) {
        return VInt(0), fmt.Errorf("type error, function '%s' expects %d args, got %d", name, len(fn.Params), len(args))
    }
    interp.callDepth++
    if interp.callDepth > maxCallDepth {
        interp.callDepth--
        return VInt(0), HaltSignal{}
    }
    defer func() { interp.callDepth-- }()
    env := NewEnv(nil)
    for i, param := range fn.Params {
        env.Set(param.Name, args[i])
    }
    err := interp.execBlock(fn.Body, env)
    if err != nil {
        if _, ok := err.(HaltSignal); ok {
            return VInt(0), err
        }
        if bs, ok := err.(BackSignal); ok {
            return bs.Value, nil
        }
        return VInt(0), err
    }
    return VInt(0), nil
}

func (interp *Interpreter) execBlock(stmts []ast.Stmt, env *Environment) error {
    for _, s := range stmts {
        err := interp.execStmt(s, env)
        if err != nil {
            return err
        }
    }
    return nil
}

func (interp *Interpreter) execStmt(s ast.Stmt, env *Environment) error {
    switch v := s.(type) {
    case *ast.LetStmt:
        val, err := interp.evalExpr(v.Init, env)
        if err != nil {
            return err
        }
        env.Set(v.Name, val)
        return nil
    case *ast.AssignStmt:
        if _, ok := env.Get(v.Name); !ok {
            return fmt.Errorf("syntax error, undefined variable '%s'", v.Name)
        }
        val, err := interp.evalExpr(v.Expr, env)
        if err != nil {
            return err
        }
        env.Set(v.Name, val)
        return nil
    case *ast.OutStmt:
        val, err := interp.evalExpr(v.Expr, env)
        if err != nil {
            return err
        }
        fmt.Println(ValueString(val))
        return nil
    case *ast.BackStmt:
        val, err := interp.evalExpr(v.Expr, env)
        if err != nil {
            return err
        }
        return BackSignal{Value: val}
    case *ast.HaltStmt:
        return HaltSignal{}
    case *ast.SkipStmt:
        return SkipSignal{}
    case *ast.IfStmt:
        return interp.execIf(v, env)
    case *ast.WhileStmt:
        return interp.execWhile(v, env)
    case *ast.ForInStmt:
        return interp.execForIn(v, env)
    case *ast.ExprStmt:
        _, err := interp.evalExpr(v.Expr, env)
        return err
    default:
        return fmt.Errorf("syntax error, unknown statement type")
    }
}

func (interp *Interpreter) execIf(s *ast.IfStmt, env *Environment) error {
    cond, err := interp.evalExpr(s.Cond, env)
    if err != nil {
        return err
    }
    if cond.Bool() {
        return interp.execBlock(s.Then, env)
    }
    for i, econd := range s.ElifConds {
        cv, err := interp.evalExpr(econd, env)
        if err != nil {
            return err
        }
        if cv.Bool() {
            return interp.execBlock(s.ElifBodies[i], env)
        }
    }
    if s.ElseBody != nil {
        return interp.execBlock(s.ElseBody, env)
    }
    return nil
}

func (interp *Interpreter) execWhile(s *ast.WhileStmt, env *Environment) error {
    for {
        cond, err := interp.evalExpr(s.Cond, env)
        if err != nil {
            return err
        }
        if !cond.Bool() {
            break
        }
        err = interp.execBlock(s.Body, env)
        if err != nil {
            if _, ok := err.(HaltSignal); ok {
                return nil  // 只跳出循环，不传播
            }
            if _, ok := err.(SkipSignal); ok {
                continue
            }
            return err
        }
    }
    return nil
}

func (interp *Interpreter) evalExpr(e ast.Expr, env *Environment) (Value, error) {
    switch v := e.(type) {
    case *ast.IntLiteral:
        return VInt(v.Value), nil
    case *ast.StringLiteral:
        return VStr(v.Value), nil
    case *ast.BoolLiteral:
        return VBool(v.Value), nil
    case *ast.Ident:
        val, ok := env.Get(v.Name)
        if !ok {
            return VInt(0), fmt.Errorf("syntax error, undefined variable '%s'", v.Name)
        }
        return val, nil
    case *ast.BinaryExpr:
        left, err := interp.evalExpr(v.Left, env)
        if err != nil {
            return VInt(0), err
        }
        right, err := interp.evalExpr(v.Right, env)
        if err != nil {
            return VInt(0), err
        }
        return interp.evalBinary(v.Op, left, right)
    case *ast.UnaryExpr:
        right, err := interp.evalExpr(v.Right, env)
        if err != nil {
            return VInt(0), err
        }
        if v.Op == token.TK_MINUS {
            if right.IsString {
                return VInt(0), fmt.Errorf("type error, cannot negate a string")
            }
            return VInt(-right.Int()), nil
        }
        return right, nil
    case *ast.CallExpr:
        var args []Value
        for _, arg := range v.Args {
            a, err := interp.evalExpr(arg, env)
            if err != nil {
                return VInt(0), err
            }
            args = append(args, a)
        }
        return interp.callFn(v.Callee, args)
    case *ast.InputExpr:
        return interp.readInput()
    default:
        return VInt(0), fmt.Errorf("syntax error, unknown expression type")
    }
}

func (interp *Interpreter) evalBinary(op token.TokenType, left, right Value) (Value, error) {
    switch op {
    case token.TK_PLUS:
        if left.IsString && right.IsString {
            return VStr(left.Str() + right.Str()), nil
        }
        if left.IsString || right.IsString {
            return VInt(0), fmt.Errorf("type error, cannot add string and number")
        }
        return VInt(left.Int() + right.Int()), nil
    case token.TK_MINUS:
        if left.IsString || right.IsString {
            return VInt(0), fmt.Errorf("type error, cannot subtract strings")
        }
        return VInt(left.Int() - right.Int()), nil
    case token.TK_STAR:
        if left.IsString || right.IsString {
            return VInt(0), fmt.Errorf("type error, cannot multiply strings")
        }
        return VInt(left.Int() * right.Int()), nil
    case token.TK_SLASH:
        if left.IsString || right.IsString {
            return VInt(0), fmt.Errorf("type error, cannot divide strings")
        }
        if right.Int() == 0 {
            return VInt(0), fmt.Errorf("type error, division by zero")
        }
        result := left.Int() / right.Int()
        if (left.Int() < 0) != (right.Int() < 0) && left.Int()%right.Int() != 0 {
            result--
        }
        return VInt(result), nil
    case token.TK_EQ:
        if left.IsString != right.IsString {
            return VBool(false), nil
        }
        if left.IsString {
            if left.Str() == right.Str() {
                return VBool(true), nil
            }
            return VBool(false), nil
        }
        if left.Int() == right.Int() {
            return VBool(true), nil
        }
        return VBool(false), nil
    case token.TK_NEQ:
        if left.IsString != right.IsString {
            return VBool(true), nil
        }
        if left.IsString {
            if left.Str() != right.Str() {
                return VBool(true), nil
            }
            return VBool(false), nil
        }
        if left.Int() != right.Int() {
            return VBool(true), nil
        }
        return VBool(false), nil
    case token.TK_LT:
        if left.IsString || right.IsString {
            return VInt(0), fmt.Errorf("type error, cannot compare strings with <")
        }
        if left.Int() < right.Int() {
            return VBool(true), nil
        }
        return VBool(false), nil
    case token.TK_GT:
        if left.IsString || right.IsString {
            return VInt(0), fmt.Errorf("type error, cannot compare strings with >")
        }
        if left.Int() > right.Int() {
            return VBool(true), nil
        }
        return VBool(false), nil
    case token.TK_LE:
        if left.IsString || right.IsString {
            return VInt(0), fmt.Errorf("type error, cannot compare strings with <=")
        }
        if left.Int() <= right.Int() {
            return VBool(true), nil
        }
        return VBool(false), nil
    case token.TK_GE:
        if left.IsString || right.IsString {
            return VInt(0), fmt.Errorf("type error, cannot compare strings with >=")
        }
        if left.Int() >= right.Int() {
            return VBool(true), nil
        }
        return VBool(false), nil
    default:
        return VInt(0), fmt.Errorf("syntax error, unknown operator")
    }
}

func (interp *Interpreter) readInput() (Value, error) {
    var s string
    _, err := fmt.Scan(&s)
    if err != nil {
        os.Exit(0)
    }
    val, err := strconv.ParseInt(s, 10, 64)
    if err != nil {
        return VInt(0), fmt.Errorf("type error, input must be an integer")
    }
    return VInt(val), nil
}

func (interp *Interpreter) execForIn(s *ast.ForInStmt, env *Environment) error {
    iterVal, err := interp.evalExpr(s.IterExpr, env)
    if err != nil {
        return err
    }
    if !iterVal.IsString {
        return fmt.Errorf("type error, 'for...in' requires a string")
    }
    for _, r := range iterVal.StrValue {
        chEnv := NewEnv(env)
        chEnv.Set(s.VarName, VStr(string(r)))
        err := interp.execBlock(s.Body, chEnv)
        if err != nil {
            if _, ok := err.(HaltSignal); ok {
                return nil  // 只跳出循环，不传播
            }
            if _, ok := err.(SkipSignal); ok {
                continue
            }
            return err
        }
    }
    return nil
}
