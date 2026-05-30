package lexer

import (
    "fmt"
    "luminalang/token"
    "unicode"
)

type Lexer struct {
    source  []rune
    tokens  []token.Token
    start   int
    current int
    line    int
    column  int
}

func New(source string) *Lexer {
    return &Lexer{
        source:  []rune(source),
        tokens:  nil,
        start:   0,
        current: 0,
        line:    1,
        column:  1,
    }
}

func (l *Lexer) Tokenize() ([]token.Token, error) {
    for !l.atEnd() {
        l.start = l.current
        if err := l.scanToken(); err != nil {
            return nil, err
        }
    }
    l.tokens = append(l.tokens, token.Token{
        Type:   token.TK_EOF,
        Lexeme: "",
        Line:   l.line,
        Column: l.column,
    })
    return l.tokens, nil
}

func (l *Lexer) atEnd() bool {
    return l.current >= len(l.source)
}

func (l *Lexer) advance() rune {
    ch := l.source[l.current]
    l.current++
    if ch == '\n' {
        l.line++
        l.column = 1
    } else {
        l.column++
    }
    return ch
}

func (l *Lexer) peek() rune {
    if l.atEnd() {
        return 0
    }
    return l.source[l.current]
}

func (l *Lexer) match(expected rune) bool {
    if l.atEnd() || l.source[l.current] != expected {
        return false
    }
    l.advance()
    return true
}

func (l *Lexer) addToken(t token.TokenType, value int64) {
    lexeme := string(l.source[l.start:l.current])
    l.tokens = append(l.tokens, token.Token{
        Type:   t,
        Lexeme: lexeme,
        Value:  value,
        Line:   l.line,
        Column: l.column - len(lexeme),
    })
}

func (l *Lexer) scanToken() error {
    ch := l.advance()
    switch ch {
    case ' ', '\t', '\r', '\n':
    case '+':
        l.addToken(token.TK_PLUS, 0)
    case '-':
        l.addToken(token.TK_MINUS, 0)
    case '*':
        l.addToken(token.TK_STAR, 0)
    case '/':
        l.addToken(token.TK_SLASH, 0)
    case '(':
        l.addToken(token.TK_LPAREN, 0)
    case ')':
        l.addToken(token.TK_RPAREN, 0)
    case '{':
        l.addToken(token.TK_LBRACE, 0)
    case '}':
        l.addToken(token.TK_RBRACE, 0)
    case ',':
        l.addToken(token.TK_COMMA, 0)
    case ':':
        if l.match(':') {
            l.addToken(token.TK_DCOLON, 0)
        } else {
            l.addToken(token.TK_COLON, 0)
        }
    case '=':
        if l.match('=') {
            l.addToken(token.TK_EQ, 0)
        } else {
            l.addToken(token.TK_ASSIGN, 0)
        }
    case '!':
        if l.match('=') {
            l.addToken(token.TK_NEQ, 0)
        } else {
            return fmt.Errorf("line %d: syntax error, unexpected character '%c'", l.line, ch)
        }
    case '<':
        if l.match('=') {
            l.addToken(token.TK_LE, 0)
        } else {
            l.addToken(token.TK_LT, 0)
        }
    case '>':
        if l.match('=') {
            l.addToken(token.TK_GE, 0)
        } else {
            l.addToken(token.TK_GT, 0)
        }
    default:
        if unicode.IsDigit(ch) {
            return l.number()
        } else if isIdentStart(ch) {
            l.identifier()
        } else {
            return fmt.Errorf("line %d: syntax error, unexpected character '%c'", l.line, ch)
        }
    }
    return nil
}

func (l *Lexer) number() error {
    for !l.atEnd() && unicode.IsDigit(l.peek()) {
        l.advance()
    }
    if !l.atEnd() && (l.peek() == '.' || l.peek() == 'e' || l.peek() == 'E') {
        return fmt.Errorf("line %d: type error, floating-point literals not supported", l.line)
    }
    val := 0
    for _, ch := range l.source[l.start:l.current] {
        val = val*10 + int(ch-'0')
    }
    l.addToken(token.TK_INT, int64(val))
    return nil
}

var keywords = map[string]token.TokenType{
    "let":   token.TK_LET,
    "fn":    token.TK_FN,
    "while": token.TK_WHILE,
    "if":    token.TK_IF,
    "elif":  token.TK_ELIF,
    "else":  token.TK_ELSE,
    "back":  token.TK_BACK,
    "out":   token.TK_OUT,
    "input": token.TK_INPUT,
    "i32":   token.TK_I32,
    "use":   token.TK_USE,
}

func (l *Lexer) identifier() {
    for !l.atEnd() && isIdentPart(l.peek()) {
        l.advance()
    }
    lexeme := string(l.source[l.start:l.current])
    if t, ok := keywords[lexeme]; ok {
        l.addToken(t, 0)
    } else {
        l.addToken(token.TK_IDENT, 0)
    }
}

func isIdentStart(ch rune) bool {
    return unicode.IsLetter(ch) || ch == '_'
}

func isIdentPart(ch rune) bool {
    return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_'
}
