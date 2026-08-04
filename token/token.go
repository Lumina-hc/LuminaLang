package token

type TokenType int

const (
	TK_EOF TokenType = iota
	TK_LET
	TK_FN
	TK_WHILE
	TK_IF
	TK_ELIF
	TK_ELSE
	TK_BACK
	TK_OUT
	TK_INPUT
	TK_HALT
	TK_SKIP
	TK_COMMENT_SL
	TK_COMMENT_ML
	TK_STRING
	TK_I32
	TK_USE
	TK_IDENT
	TK_INT
	TK_PLUS
	TK_MINUS
	TK_STAR
	TK_SLASH
	TK_EQ
	TK_NEQ
	TK_LT
	TK_GT
	TK_LE
	TK_GE
	TK_ASSIGN
	TK_LPAREN
	TK_RPAREN
	TK_LBRACE
	TK_RBRACE
	TK_COMMA
	TK_COLON
	TK_DCOLON
	TK_BOOL
	TK_TRUE
	TK_FALSE
	TK_FOR
	TK_IN
	TK_STRING_TYPE
)

type Token struct {
	Type   TokenType
	Lexeme string
	Value  int64
	Line   int
	Column int
}

func (t Token) String() string {
	return t.Lexeme
}
