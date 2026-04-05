import sys

class Token:
    FN     = "FN"
    LET    = "LET"
    BACK   = "BACK"
    IF     = "IF"
    ELIF   = "ELIF"
    ELSE   = "ELSE"
    WHILE  = "WHILE"
    OUT    = "OUT"
    ID     = "ID"
    TYPE   = "TYPE"
    INT    = "INT"
    LPAREN = "("
    RPAREN = ")"
    LBRACE = "{"
    RBRACE = "}"
    COLON  = ":"
    COMMA  = ","
    PLUS   = "+"
    MINUS  = "-"
    MUL    = "*"
    DIV    = "/"
    EQ     = "=="
    NEQ    = "!="
    LT     = "<"
    GT     = ">"
    LTE    = "<="
    GTE    = ">="
    ASSIGN = "="
    EOF    = "EOF"

    def __init__(self, type_, val):
        self.type = type_
        self.val  = val

class Lexer:
    def __init__(self, text):
        self.text = text
        self.pos  = 0
        self.cur  = text[self.pos] if text else None

    def advance(self):
        self.pos += 1
        self.cur = self.text[self.pos] if self.pos < len(self.text) else None

    def skip_whitespace(self):
        while self.cur and self.cur.isspace():
            self.advance()

    def read_ident(self):
        s = self.pos
        while self.cur and (self.cur.isalnum() or self.cur == "_"):
            self.advance()
        return self.text[s:self.pos]

    def read_int(self):
        s = self.pos
        while self.cur and self.cur.isdigit():
            self.advance()
        return int(self.text[s:self.pos])

    def next(self):
        while self.cur:
            if self.cur.isspace():
                self.skip_whitespace()
                continue
            if self.cur.isalpha() or self.cur == "_":
                v = self.read_ident()
                if   v == "fn":    return Token(Token.FN, v)
                elif v == "let":   return Token(Token.LET, v)
                elif v == "back":  return Token(Token.BACK, v)
                elif v == "if":    return Token(Token.IF, v)
                elif v == "elif":  return Token(Token.ELIF, v)
                elif v == "else":  return Token(Token.ELSE, v)
                elif v == "while": return Token(Token.WHILE, v)
                elif v == "out":   return Token(Token.OUT, v)
                elif v == "i32":   return Token(Token.TYPE, v)
                else:              return Token(Token.ID, v)
            if self.cur.isdigit():
                return Token(Token.INT, self.read_int())
            if self.cur == '(': self.advance(); return Token(Token.LPAREN, '(')
            if self.cur == ')': self.advance(); return Token(Token.RPAREN, ')')
            if self.cur == '{': self.advance(); return Token(Token.LBRACE, '{')
            if self.cur == '}': self.advance(); return Token(Token.RBRACE, '}')
            if self.cur == ':': self.advance(); return Token(Token.COLON, ':')
            if self.cur == ',': self.advance(); return Token(Token.COMMA, ',')
            if self.cur == '+': self.advance(); return Token(Token.PLUS, '+')
            if self.cur == '-': self.advance(); return Token(Token.MINUS, '-')
            if self.cur == '*': self.advance(); return Token(Token.MUL, '*')
            if self.cur == '/': self.advance(); return Token(Token.DIV, '/')
            if self.cur == '=':
                self.advance()
                if self.cur == '=':
                    self.advance()
                    return Token(Token.EQ, '==')
                else:
                    return Token(Token.ASSIGN, '=')
            if self.cur == '!':
                self.advance()
                self.advance()
                return Token(Token.NEQ, '!=')
            if self.cur == '<':
                self.advance()
                if self.cur == '=':
                    self.advance()
                    return Token(Token.LTE, '<=')
                else:
                    return Token(Token.LT, '<')
            if self.cur == '>':
                self.advance()
                if self.cur == '=':
                    self.advance()
                    return Token(Token.GTE, '>=')
                else:
                    return Token(Token.GT, '>')
            raise Exception("语法错误")
        return Token(Token.EOF, None)

class Parser:
    def __init__(self, lexer):
        self.lex = lexer
        self.tok = self.lex.next()

    def eat(self, t):
        if self.tok.type == t:
            self.tok = self.lex.next()
        else:
            raise Exception("语法错误")

    def parse_expr(self):
        return self.parse_comp()

    def parse_comp(self):
        node = self.parse_term()
        while self.tok.type in (Token.EQ, Token.NEQ, Token.LT, Token.GT, Token.LTE, Token.GTE):
            op = self.tok.type
            self.eat(op)
            node = ("binop", op, node, self.parse_term())
        return node

    def parse_term(self):
        node = self.parse_fact()
        while self.tok.type in (Token.PLUS, Token.MINUS):
            op = self.tok.type
            self.eat(op)
            node = ("binop", op, node, self.parse_fact())
        return node

    def parse_fact(self):
        if self.tok.type == Token.INT:
            node = ("int", self.tok.val)
            self.eat(Token.INT)
            return node
        if self.tok.type == Token.ID:
            name = self.tok.val
            self.eat(Token.ID)
            if self.tok.type == Token.LPAREN:
                self.eat(Token.LPAREN)
                args = []
                if self.tok.type != Token.RPAREN:
                    args.append(self.parse_expr())
                    while self.tok.type == Token.COMMA:
                        self.eat(Token.COMMA)
                        args.append(self.parse_expr())
                self.eat(Token.RPAREN)
                if name == "input":
                    return ("input",)
                return ("call", name, args)
            else:
                return ("var", name)
        if self.tok.type == Token.LPAREN:
            self.eat(Token.LPAREN)
            node = self.parse_expr()
            self.eat(Token.RPAREN)
            return node
        raise Exception("语法错误")

    def parse_stmt(self):
        if self.tok.type == Token.LET:
            self.eat(Token.LET)
            name = self.tok.val
            self.eat(Token.ID)
            self.eat(Token.COLON)
            ty = self.tok.val
            self.eat(Token.TYPE)
            self.eat(Token.ASSIGN)
            val = self.parse_expr()
            return ("let", name, ty, val)

        if self.tok.type == Token.ID:
            name = self.tok.val
            self.eat(Token.ID)
            if self.tok.type == Token.ASSIGN:
                self.eat(Token.ASSIGN)
                val = self.parse_expr()
                return ("set", name, val)
            else:
                return self.parse_expr()

        if self.tok.type == Token.IF:
            self.eat(Token.IF)
            cond = self.parse_expr()
            self.eat(Token.LBRACE)
            body = self.parse_stmts()
            self.eat(Token.RBRACE)
            elifs = []
            while self.tok.type == Token.ELIF:
                self.eat(Token.ELIF)
                c = self.parse_expr()
                self.eat(Token.LBRACE)
                b = self.parse_stmts()
                self.eat(Token.RBRACE)
                elifs.append((c, b))
            els = None
            if self.tok.type == Token.ELSE:
                self.eat(Token.ELSE)
                self.eat(Token.LBRACE)
                els = self.parse_stmts()
                self.eat(Token.RBRACE)
            return ("if", cond, body, elifs, els)

        if self.tok.type == Token.WHILE:
            self.eat(Token.WHILE)
            cond = self.parse_expr()
            self.eat(Token.LBRACE)
            body = self.parse_stmts()
            self.eat(Token.RBRACE)
            return ("while", cond, body)

        if self.tok.type == Token.OUT:
            self.eat(Token.OUT)
            val = self.parse_expr()
            return ("out", val)

        if self.tok.type == Token.BACK:
            self.eat(Token.BACK)
            e = self.parse_expr()
            return ("back", e)

        return self.parse_expr()

    def parse_stmts(self):
        stmts = []
        while self.tok.type not in (Token.RBRACE, Token.EOF):
            stmts.append(self.parse_stmt())
        return stmts

    def parse_fn(self):
        self.eat(Token.FN)
        name = self.tok.val
        self.eat(Token.ID)
        self.eat(Token.LPAREN)
        params = []
        while self.tok.type != Token.RPAREN:
            pname = self.tok.val
            self.eat(Token.ID)
            self.eat(Token.COLON)
            ptype = self.tok.val
            self.eat(Token.TYPE)
            params.append((pname, ptype))
            if self.tok.type == Token.COMMA:
                self.eat(Token.COMMA)
        self.eat(Token.RPAREN)
        rtype = self.tok.val
        self.eat(Token.TYPE)
        self.eat(Token.LBRACE)
        body = self.parse_stmts()
        self.eat(Token.RBRACE)
        return ("fn", name, params, rtype, body)

    def parse(self):
        fns = []
        while self.tok.type == Token.FN:
            fns.append(self.parse_fn())
        if self.tok.type != Token.EOF:
            raise Exception("语法错误")
        return fns

class Interpreter:
    def __init__(self):
        self.fns = {}

    def check(self, v, t):
        if t == "i32" and not isinstance(v, int):
            raise Exception("类型错误")

    def eval(self, node, env):
        if node[0] == "int":
            return node[1]
        if node[0] == "var":
            return env[node[1]]
        if node[0] == "input":
            return int(input())
        if node[0] == "binop":
            _, op, a, b = node
            x = self.eval(a, env)
            y = self.eval(b, env)
            if op == Token.PLUS:  return x + y
            if op == Token.MINUS: return x - y
            if op == Token.MUL:   return x * y
            if op == Token.DIV:   return x // y
            if op == Token.EQ:    return 1 if x == y else 0
            if op == Token.NEQ:   return 1 if x != y else 0
            if op == Token.LT:    return 1 if x <  y else 0
            if op == Token.GT:    return 1 if x >  y else 0
            if op == Token.LTE:   return 1 if x <= y else 0
            if op == Token.GTE:   return 1 if x >= y else 0
        if node[0] == "call":
            _, name, args = node
            fn = self.fns[name]
            _, _, params, rt, body = fn
            e = {}
            for (pn, pt), v in zip(params, [self.eval(a, env) for a in args]):
                self.check(v, pt)
                e[pn] = v
            return self.exec(body, e)
        return 0

    def exec(self, stmts, env):
        for s in stmts:
            if s[0] == "let":
                _, n, t, v = s
                val = self.eval(v, env)
                self.check(val, t)
                env[n] = val
            elif s[0] == "set":
                _, n, v = s
                env[n] = self.eval(v, env)
            elif s[0] == "if":
                _, c, b, elifs, els = s
                if self.eval(c, env) == 1:
                    res = self.exec(b, env)
                    if res is not None: return res
                    continue
                ok = False
                for ec, eb in elifs:
                    if self.eval(ec, env) == 1:
                        res = self.exec(eb, env)
                        if res is not None: return res
                        ok = True
                        break
                if ok: continue
                if els:
                    res = self.exec(els, env)
                    if res is not None: return res
            elif s[0] == "while":
                _, cond, body = s
                while self.eval(cond, env) == 1:
                    res = self.exec(body, env)
                    if res is not None:
                        return res
            elif s[0] == "out":
                val = self.eval(s[1], env)
                print(val)
            elif s[0] == "back":
                return self.eval(s[1], env)
            else:
                self.eval(s, env)
        return None

    def run(self, fns):
        for fn in fns:
            self.fns[fn[1]] = fn
        if "main" not in self.fns:
            raise Exception("未找到main")
        res = self.exec(self.fns["main"][4], {})

def main():
    try:
        with open(sys.argv[1], "r", encoding="utf-8") as f:
            code = f.read()
        ast = Parser(Lexer(code)).parse()
        Interpreter().run(ast)
    except KeyboardInterrupt:
        return
    except Exception as e:
        print(e)

if __name__ == "__main__":
    main()