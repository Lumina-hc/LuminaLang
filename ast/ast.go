package ast

import "luminalang/token"

type Expr interface {
	ExprNode()
}

type IntLiteral struct {
	Value int64
}

func (IntLiteral) ExprNode() {}

type StringLiteral struct {
	Value string
}

func (StringLiteral) ExprNode() {}

type BoolLiteral struct {
	Value bool
}

func (BoolLiteral) ExprNode() {}

type Ident struct {
	Name string
}

func (Ident) ExprNode() {}

type BinaryExpr struct {
	Left  Expr
	Op    token.TokenType
	Right Expr
}

func (BinaryExpr) ExprNode() {}

type UnaryExpr struct {
	Op    token.TokenType
	Right Expr
}

func (UnaryExpr) ExprNode() {}

type CallExpr struct {
	Callee string
	Args   []Expr
}

func (CallExpr) ExprNode() {}

type InputExpr struct{}

func (InputExpr) ExprNode() {}

type Stmt interface {
	StmtNode()
}

type LetStmt struct {
	Name string
	Init Expr
}

func (LetStmt) StmtNode() {}

type AssignStmt struct {
	Name string
	Expr Expr
}

func (AssignStmt) StmtNode() {}

type OutStmt struct {
	Expr Expr
}

func (OutStmt) StmtNode() {}

type BackStmt struct {
	Expr Expr
}

func (BackStmt) StmtNode() {}

type HaltStmt struct{}

func (HaltStmt) StmtNode() {}

type SkipStmt struct{}

func (SkipStmt) StmtNode() {}

type IfStmt struct {
	Cond       Expr
	Then       []Stmt
	ElifConds  []Expr
	ElifBodies [][]Stmt
	ElseBody   []Stmt
}

func (IfStmt) StmtNode() {}

type WhileStmt struct {
	Cond Expr
	Body []Stmt
}

func (WhileStmt) StmtNode() {}

type ForInStmt struct {
	VarName  string
	IterExpr Expr
	Body     []Stmt
}

func (ForInStmt) StmtNode() {}

type ExprStmt struct {
	Expr Expr
}

func (ExprStmt) StmtNode() {}

type Decl interface {
	DeclNode()
}

type FnDecl struct {
	Name    string
	Params  []Param
	RetType string
	Body    []Stmt
}

func (FnDecl) DeclNode() {}

type Param struct {
	Name string
	Type string
}

type UseDecl struct {
	Module   string
	Items    []string
	Wildcard bool
}

func (UseDecl) DeclNode() {}

type Program struct {
	Imports   []UseDecl
	Functions []FnDecl
}
