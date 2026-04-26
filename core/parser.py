from core.lexer import Lexer, TokenType

class Parser:
    def __init__(self, lexer: Lexer):
        self.lex = lexer
        self.tok = self.lex.next_token()

    def eat(self, type_):
        if self.tok.type == type_:
            self.tok = self.lex.next_token()
        else:
            raise Exception(f"[语法错误] 期望 {type_} 得到 {self.tok.type}")

    def parse(self):
        fns = []
        while self.tok.type != TokenType.EOF:
            if self.tok.type == TokenType.USE:
                fns.append(self.use_stmt())
            else:
                fns.append(self.fn())
        return fns

    def use_stmt(self):
        self.eat(TokenType.USE)
        mod = self.tok.value
        self.eat(TokenType.IDENT)
        self.eat(TokenType.DOUBLE_COLON)
        name = self.tok.value
        self.eat(TokenType.IDENT)
        return ("use", mod, name)

    def fn(self):
        self.eat(TokenType.FN)
        name = self.tok.value
        self.eat(TokenType.IDENT)
        self.eat(TokenType.LPAREN)
        params = []
        while self.tok.type != TokenType.RPAREN:
            pname = self.tok.value
            self.eat(TokenType.IDENT)
            self.eat(TokenType.COLON)
            ptype = self.tok.value
            self.eat(TokenType.IDENT)
            params.append((pname, ptype))
            if self.tok.type == TokenType.COMMA:
                self.eat(TokenType.COMMA)
        self.eat(TokenType.RPAREN)
        rtype = self.tok.value
        self.eat(TokenType.IDENT)
        self.eat(TokenType.LBRACE)
        body = self.stmts()
        self.eat(TokenType.RBRACE)
        return ("fn", name, params, rtype, body)

    def stmts(self):
        stmts = []
        while self.tok.type not in (TokenType.RBRACE, TokenType.EOF):
            stmts.append(self.stmt())
        return stmts

    def stmt(self):
        if self.tok.type == TokenType.LET:
            self.eat(TokenType.LET)
            name = self.tok.value
            self.eat(TokenType.IDENT)
            self.eat(TokenType.COLON)
            ty = self.tok.value
            self.eat(TokenType.IDENT)
            self.eat(TokenType.ASSIGN)
            val = self.expr()
            return ("vdef", name, ty, val)

        if self.tok.type == TokenType.IF:
            self.eat(TokenType.IF)
            cond = self.expr()
            self.eat(TokenType.LBRACE)
            body = self.stmts()
            self.eat(TokenType.RBRACE)
            else_body = []
            if self.tok.type == TokenType.ELSE:
                self.eat(TokenType.ELSE)
                self.eat(TokenType.LBRACE)
                else_body = self.stmts()
                self.eat(TokenType.RBRACE)
            return ("if", cond, body, else_body, None)

        if self.tok.type == TokenType.TRY:
            self.eat(TokenType.TRY)
            self.eat(TokenType.LBRACE)
            try_body = self.stmts()
            self.eat(TokenType.RBRACE)

            self.eat(TokenType.EXCEPT)
            self.eat(TokenType.LBRACE)
            except_body = self.stmts()
            self.eat(TokenType.RBRACE)

            finally_body = []
            if self.tok.type == TokenType.FINALLY:
                self.eat(TokenType.FINALLY)
                self.eat(TokenType.LBRACE)
                finally_body = self.stmts()
                self.eat(TokenType.RBRACE)
            return ("try", try_body, except_body, finally_body)

        if self.tok.type == TokenType.OUT:
            self.eat(TokenType.OUT)
            return ("print", self.expr())

        return ("expr", self.expr())

    def expr(self): return self.or_expr()
    def or_expr(self):
        node = self.and_expr()
        while self.tok.type == TokenType.OR:
            self.eat(TokenType.OR)
            node = ("op", TokenType.OR, node, self.and_expr())
        return node
    def and_expr(self):
        node = self.cmp()
        while self.tok.type == TokenType.AND:
            self.eat(TokenType.AND)
            node = ("op", TokenType.AND, node, self.cmp())
        return node
    def cmp(self):
        node = self.term()
        while self.tok.type in (TokenType.EQ, TokenType.NEQ, TokenType.LT, TokenType.GT):
            op = self.tok.type
            self.eat(op)
            node = ("op", op, node, self.term())
        return node
    def term(self):
        node = self.factor()
        while self.tok.type in (TokenType.PLUS, TokenType.MINUS, TokenType.MUL, TokenType.DIV, TokenType.MOD):
            op = self.tok.type
            self.eat(op)
            node = ("op", op, node, self.factor())
        return node
    def factor(self):
        if self.tok.type == TokenType.NUMBER:
            v = self.tok.value
            self.eat(TokenType.NUMBER)
            return ("num", v)
        if self.tok.type == TokenType.STRING:
            v = self.tok.value
            self.eat(TokenType.STRING)
            return ("str", v)
        if self.tok.type == TokenType.IDENT:
            name = self.tok.value
            self.eat(TokenType.IDENT)
            if self.tok.type == TokenType.LPAREN:
                self.eat(TokenType.LPAREN)
                args = []
                if self.tok.type != TokenType.RPAREN:
                    args.append(self.expr())
                    while self.tok.type == TokenType.COMMA:
                        self.eat(TokenType.COMMA)
                        args.append(self.expr())
                self.eat(TokenType.RPAREN)
                return ("call", name, args)
            return ("var", name)
        if self.tok.type == TokenType.LPAREN:
            self.eat(TokenType.LPAREN)
            e = self.expr()
            self.eat(TokenType.RPAREN)
            return e
        raise Exception(f"[表达式错误] 无法解析: {self.tok.type}")