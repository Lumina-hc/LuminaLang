package runtime

import (
    "fmt"
    "luminalang/ast"
    "luminalang/token"
    "os"
    "strconv"
)

type Value = int64

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
    return 0, false
}

func (e *Environment) Set(name string, val Value) {
    e.vars[name] = val
}

type BackSignal struct {
    Value Value
}

func (BackSignal) Error() string { return "__back__" }

type Interpreter struct {
    functions map[string]ast.FnDecl
    builtins  map[string]BuiltinFn
}

func NewInterpreter() *Interpreter {
    return &Interpreter{
        functions: make(map[string]ast.FnDecl),
        builtins:  make(map[string]BuiltinFn),
    }
}

func (interp *Interpreter) RegisterBuiltin(name string, fn BuiltinFn) {
    interp.builtins[name] = fn
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
        return 1, err
    }
    return int(result), nil
}

func (interp *Interpreter) callFn(name string, args []Value) (Value, error) {
    if bfn, ok := interp.builtins[name]; ok {
        return bfn(args)
    }
    fn, ok := interp.functions[name]
    if !ok {
        return 0, fmt.Errorf("type error, undefined function '%s'", name)
    }
    if len(args) != len(fn.Params) {
        return 0, fmt.Errorf("type error, function '%s' expects %d args, got %d", name, len(fn.Params), len(args))
    }
    env := NewEnv(nil)
    for i, param := range fn.Params {
        env.Set(param.Name, args[i])
    }
    err := interp.execBlock(fn.Body, env)
    if err != nil {
        if bs, ok := err.(BackSignal); ok {
            return bs.Value, nil
        }
        return 0, err
    }
    return 0, nil
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
        fmt.Println(val)
        return nil
    case *ast.BackStmt:
        val, err := interp.evalExpr(v.Expr, env)
        if err != nil {
            return err
        }
        return BackSignal{Value: val}
    case *ast.IfStmt:
        return interp.execIf(v, env)
    case *ast.WhileStmt:
        return interp.execWhile(v, env)
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
    if cond != 0 {
        return interp.execBlock(s.Then, env)
    }
    for i, econd := range s.ElifConds {
        cv, err := interp.evalExpr(econd, env)
        if err != nil {
            return err
        }
        if cv != 0 {
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
        if cond == 0 {
            break
        }
        err = interp.execBlock(s.Body, env)
        if err != nil {
            return err
        }
    }
    return nil
}

func (interp *Interpreter) evalExpr(e ast.Expr, env *Environment) (Value, error) {
    switch v := e.(type) {
    case *ast.IntLiteral:
        return v.Value, nil
    case *ast.Ident:
        val, ok := env.Get(v.Name)
        if !ok {
            return 0, fmt.Errorf("syntax error, undefined variable '%s'", v.Name)
        }
        return val, nil
    case *ast.BinaryExpr:
        left, err := interp.evalExpr(v.Left, env)
        if err != nil {
            return 0, err
        }
        right, err := interp.evalExpr(v.Right, env)
        if err != nil {
            return 0, err
        }
        return interp.evalBinary(v.Op, left, right)
    case *ast.UnaryExpr:
        right, err := interp.evalExpr(v.Right, env)
        if err != nil {
            return 0, err
        }
        if v.Op == token.TK_MINUS {
            return -right, nil
        }
        return right, nil
    case *ast.CallExpr:
        var args []Value
        for _, arg := range v.Args {
            a, err := interp.evalExpr(arg, env)
            if err != nil {
                return 0, err
            }
            args = append(args, a)
        }
        return interp.callFn(v.Callee, args)
    case *ast.InputExpr:
        return interp.readInput()
    default:
        return 0, fmt.Errorf("syntax error, unknown expression type")
    }
}

func (interp *Interpreter) evalBinary(op token.TokenType, left, right Value) (Value, error) {
    switch op {
    case token.TK_PLUS:
        return left + right, nil
    case token.TK_MINUS:
        return left - right, nil
    case token.TK_STAR:
        return left * right, nil
    case token.TK_SLASH:
        if right == 0 {
            return 0, fmt.Errorf("type error, division by zero")
        }
        result := left / right
        if (left < 0) != (right < 0) && left%right != 0 {
            result--
        }
        return result, nil
    case token.TK_EQ:
        if left == right {
            return 1, nil
        }
        return 0, nil
    case token.TK_NEQ:
        if left != right {
            return 1, nil
        }
        return 0, nil
    case token.TK_LT:
        if left < right {
            return 1, nil
        }
        return 0, nil
    case token.TK_GT:
        if left > right {
            return 1, nil
        }
        return 0, nil
    case token.TK_LE:
        if left <= right {
            return 1, nil
        }
        return 0, nil
    case token.TK_GE:
        if left >= right {
            return 1, nil
        }
        return 0, nil
    default:
        return 0, fmt.Errorf("syntax error, unknown operator")
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
        return 0, fmt.Errorf("type error, input must be an integer")
    }
    return val, nil
}
