package token

type TokenType int

const (
    TK_EOF    TokenType = iota
    TK_LET
    TK_FN
    TK_WHILE
    TK_IF
    TK_ELIF
    TK_ELSE
    TK_BACK
    TK_OUT
    TK_INPUT
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
)

type Token struct {
    Type   TokenType
    Lexeme string
    Value  int64
    Line   int
    Column int
}

func (t Token) String() string {
    if t.Type == TK_INT {
        return t.Lexeme
    }
    return t.Lexeme
}
