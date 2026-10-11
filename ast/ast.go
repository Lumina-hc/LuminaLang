package ast

import "luminalang/token"

type Expr interface {
	ExprNode()
}

type IntLiteral struct {
	Value int64
}

func (IntLiteral) ExprNode() {}

type FloatLiteral struct {
	Value float64
}

func (FloatLiteral) ExprNode() {}

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

type SliceLiteral struct {
	PosInfo
	Elems []Expr
	Type  string
}

func (SliceLiteral) ExprNode() {}

type IndexExpr struct {
	PosInfo
	Target Expr
	Index  Expr
}

func (IndexExpr) ExprNode() {}

type InputExpr struct{}

func (InputExpr) ExprNode() {}

type PosInfo struct {
	Line   int
	Column int
}

func (p PosInfo) Pos() (int, int) { return p.Line, p.Column }

type Stmt interface {
	StmtNode()
}

type LetStmt struct {
	PosInfo
	Name string
	Type string
	Init Expr
}

func (LetStmt) StmtNode() {}

type AssignStmt struct {
	PosInfo
	Name string
	Expr Expr
}

func (AssignStmt) StmtNode() {}

type OutStmt struct {
	PosInfo
	Expr Expr
}

func (OutStmt) StmtNode() {}

type BackStmt struct {
	PosInfo
	Expr Expr
}

func (BackStmt) StmtNode() {}

type HaltStmt struct {
	PosInfo
}

func (HaltStmt) StmtNode() {}

type SkipStmt struct {
	PosInfo
}

func (SkipStmt) StmtNode() {}

type IfStmt struct {
	PosInfo
	Cond       Expr
	Then       []Stmt
	ElifConds  []Expr
	ElifBodies [][]Stmt
	ElseBody   []Stmt
}

func (IfStmt) StmtNode() {}

type WhileStmt struct {
	PosInfo
	Cond Expr
	Body []Stmt
}

func (WhileStmt) StmtNode() {}

type ForInStmt struct {
	PosInfo
	VarName  string
	IterExpr Expr
	Body     []Stmt
}

func (ForInStmt) StmtNode() {}

type ExprStmt struct {
	PosInfo
	Expr Expr
}

func (ExprStmt) StmtNode() {}

type IndexAssignStmt struct {
	PosInfo
	Name  string
	Index Expr
	Value Expr
}

func (IndexAssignStmt) StmtNode() {}

type Decl interface {
	DeclNode()
}

type FnDecl struct {
	PosInfo
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
	Items    []ImportItem
	Wildcard bool
}

type ImportItem struct {
	Name  string
	Alias string
}

func (UseDecl) DeclNode() {}

type Program struct {
	Imports   []UseDecl
	Functions []FnDecl
}
