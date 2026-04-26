from typing import Any

class TokenType:
    FN = "FN"
    LET = "LET"
    IF = "IF"
    ELSE = "ELSE"
    TRY = "TRY"
    EXCEPT = "EXCEPT"
    FINALLY = "FINALLY"
    OUT = "OUT"
    USE = "USE"
    IDENT = "IDENT"
    NUMBER = "NUMBER"
    STRING = "STRING"
    LPAREN = "LPAREN"
    RPAREN = "RPAREN"
    LBRACE = "LBRACE"
    RBRACE = "RBRACE"
    PLUS = "PLUS"
    MINUS = "MINUS"
    MUL = "MUL"
    DIV = "DIV"
    MOD = "MOD"
    EQ = "EQ"
    NEQ = "NEQ"
    LT = "LT"
    GT = "GT"
    AND = "AND"
    OR = "OR"
    ASSIGN = "ASSIGN"
    COLON = "COLON"
    DOUBLE_COLON = "DOUBLE_COLON"
    COMMA = "COMMA"
    EOF = "EOF"

class Token:
    def __init__(self, type_, value):
        self.type = type_
        self.value = value

class Lexer:
    def __init__(self, code):
        self.code = code
        self.pos = 0
        self.current_char = code[self.pos] if code else None

    def advance(self):
        self.pos += 1
        self.current_char = self.code[self.pos] if self.pos < len(self.code) else None

    def skip_whitespace(self):
        while self.current_char and self.current_char.isspace():
            self.advance()

    def skip_comment(self):
        if self.current_char == "/":
            self.advance()
            if self.current_char == "/":
                self.advance()
                while self.current_char and self.current_char != "\n":
                    self.advance()

    def read_ident(self):
        res = ""
        while self.current_char and (self.current_char.isalnum() or self.current_char == "_"):
            res += self.current_char
            self.advance()
        return res

    def read_number(self):
        res = ""
        while self.current_char and self.current_char.isdigit():
            res += self.current_char
            self.advance()
        return int(res)

    def read_string(self):
        self.advance()
        res = ""
        while self.current_char and self.current_char != '"':
            res += self.current_char
            self.advance()
        self.advance()
        return res

    def next_token(self):
        while self.current_char:
            self.skip_whitespace()
            self.skip_comment()
            if not self.current_char:
                break

            if self.current_char.isalpha():
                ident = self.read_ident()
                kw = {
                    "fn": TokenType.FN,
                    "let": TokenType.LET,
                    "if": TokenType.IF,
                    "else": TokenType.ELSE,
                    "try": TokenType.TRY,
                    "except": TokenType.EXCEPT,
                    "finally": TokenType.FINALLY,
                    "out": TokenType.OUT,
                    "use": TokenType.USE,
                    "and": TokenType.AND,
                    "or": TokenType.OR
                }
                return Token(kw.get(ident, TokenType.IDENT), ident)

            if self.current_char.isdigit():
                return Token(TokenType.NUMBER, self.read_number())
            if self.current_char == '"':
                return Token(TokenType.STRING, self.read_string())

            if self.current_char == "(": self.advance(); return Token(TokenType.LPAREN, "(")
            if self.current_char == ")": self.advance(); return Token(TokenType.RPAREN, ")")
            if self.current_char == "{": self.advance(); return Token(TokenType.LBRACE, "{")
            if self.current_char == "}": self.advance(); return Token(TokenType.RBRACE, "}")
            if self.current_char == "+": self.advance(); return Token(TokenType.PLUS, "+")
            if self.current_char == "-": self.advance(); return Token(TokenType.MINUS, "-")
            if self.current_char == "*": self.advance(); return Token(TokenType.MUL, "*")
            if self.current_char == "/": self.advance(); return Token(TokenType.DIV, "/")
            if self.current_char == "%": self.advance(); return Token(TokenType.MOD, "%")
            if self.current_char == ",": self.advance(); return Token(TokenType.COMMA, ",")

            if self.current_char == "=":
                self.advance()
                if self.current_char == "=":
                    self.advance(); return Token(TokenType.EQ, "==")
                return Token(TokenType.ASSIGN, "=")
            if self.current_char == "!":
                self.advance()
                if self.current_char == "=":
                    self.advance(); return Token(TokenType.NEQ, "!=")
            if self.current_char == "<": self.advance(); return Token(TokenType.LT, "<")
            if self.current_char == ">": self.advance(); return Token(TokenType.GT, ">")
            if self.current_char == ":":
                self.advance()
                if self.current_char == ":":
                    self.advance(); return Token(TokenType.DOUBLE_COLON, "::")
                return Token(TokenType.COLON, ":")
            self.advance()
        return Token(TokenType.EOF, None)