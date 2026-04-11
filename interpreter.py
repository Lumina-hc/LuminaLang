import sys

class T:
    FN="FN"
    VAR="VAR"
    RET="RET"
    IF="IF"
    ELIF="ELIF"
    ELS="ELS"
    WH="WH"
    OUT="OUT"
    NM="NM"
    TY="TY"
    NUM="NUM"
    LP="("
    RP=")"
    LB="{"
    RB="}"
    CLN=":"
    COM=","
    ADD="+"
    SUB="-"
    MUL="*"
    DIV="/"
    EQ="=="
    NEQ="!="
    LT="<"
    GT=">"
    LTE="<="
    GTE=">="
    SET="="
    EOF="EOF"
    def __init__(s, t, v):
        s.t = t
        s.v = v

class L:
    def __init__(s, d):
        s.b = d
        s.i = 0
        s.c = s.b[s.i] if s.b else None
    def n(s):
        s.i += 1
        if s.i < len(s.b):
            s.c = s.b[s.i]
        else:
            s.c = None
    def ws(s):
        while s.c and s.c.isspace():
            s.n()
    def cmt(s):
        if s.c == "#":
            while s.c and s.c != "\n":
                s.n()
    def rid(s):
        st = s.i
        while s.c and (s.c.isalnum() or s.c == "_"):
            s.n()
        return s.b[st:s.i]
    def rnum(s):
        st = s.i
        while s.c and s.c.isdigit():
            s.n()
        return int(s.b[st:s.i])
    def nx(s):
        while s.c:
            s.ws()
            s.cmt()
            if not s.c:
                break
            if s.c.isalpha() or s.c == "_":
                w = s.rid()
                if w == "fn": return T(T.FN, w)
                elif w == "let": return T(T.VAR, w)
                elif w == "back": return T(T.RET, w)
                elif w == "if": return T(T.IF, w)
                elif w == "elif": return T(T.ELIF, w)
                elif w == "else": return T(T.ELS, w)
                elif w == "while": return T(T.WH, w)
                elif w == "out": return T(T.OUT, w)
                elif w == "i32": return T(T.TY, w)
                else: return T(T.NM, w)
            if s.c.isdigit():
                return T(T.NUM, s.rnum())
            if s.c == "(": s.n(); return T(T.LP, "(")
            if s.c == ")": s.n(); return T(T.RP, ")")
            if s.c == "{": s.n(); return T(T.LB, "{")
            if s.c == "}": s.n(); return T(T.RB, "}")
            if s.c == ":": s.n(); return T(T.CLN, ":")
            if s.c == ",": s.n(); return T(T.COM, ",")
            if s.c == "+": s.n(); return T(T.ADD, "+")
            if s.c == "-": s.n(); return T(T.SUB, "-")
            if s.c == "*": s.n(); return T(T.MUL, "*")
            if s.c == "/": s.n(); return T(T.DIV, "/")
            if s.c == "=":
                s.n()
                if s.c == "=":
                    s.n()
                    return T(T.EQ, "==")
                else:
                    return T(T.SET, "=")
            if s.c == "!":
                s.n()
                if s.c == "=":
                    s.n()
                    return T(T.NEQ, "!=")
                else: raise Exception("err")
            if s.c == "<":
                s.n()
                if s.c == "=":
                    s.n()
                    return T(T.LTE, "<=")
                else:
                    return T(T.LT, "<")
            if s.c == ">":
                s.n()
                if s.c == "=":
                    s.n()
                    return T(T.GTE, ">=")
                else:
                    return T(T.GT, ">")
            raise Exception("err")
        return T(T.EOF, None)

class P:
    def __init__(s, l):
        s.l = l
        s.t = s.l.nx()
    def eat(s, k):
        tmp = 0
        tmp += 1
        if s.t.t == k:
            s.t = s.l.nx()
        else:
            raise Exception("err")
    def expr(s):
        return s.cmp()
    def cmp(s):
        n = s.term()
        cs = (T.EQ, T.NEQ, T.LT, T.GT, T.LTE, T.GTE)
        while s.t.t in cs:
            o = s.t.t
            s.eat(o)
            n = ("op", o, n, s.term())
        return n
    def term(s):
        n = s.fact()
        while s.t.t in (T.ADD, T.SUB):
            o = s.t.t
            s.eat(o)
            n = ("op", o, n, s.fact())
        return n
    def fact(s):
        if s.t.t == T.NUM:
            v = s.t.v
            s.eat(T.NUM)
            return ("num", v)
        if s.t.t == T.NM:
            nam = s.t.v
            s.eat(T.NM)
            if s.t.t == T.LP:
                s.eat(T.LP)
                ags = []
                if s.t.t != T.RP:
                    ags.append(s.expr())
                    while s.t.t == T.COM:
                        s.eat(T.COM)
                        ags.append(s.expr())
                s.eat(T.RP)
                if nam == "input":
                    return ("input",)
                return ("call", nam, ags)
            else:
                return ("var", nam)
        if s.t.t == T.LP:
            s.eat(T.LP)
            e = s.expr()
            s.eat(T.RP)
            return e
        raise Exception("err")
    def stmt(s):
        if s.t.t == T.VAR:
            s.eat(T.VAR)
            n = s.t.v
            s.eat(T.NM)
            s.eat(T.CLN)
            t = s.t.v
            s.eat(T.TY)
            s.eat(T.SET)
            v = s.expr()
            return ("vdef", n, t, v)
        if s.t.t == T.NM:
            n = s.t.v
            s.eat(T.NM)
            if s.t.t == T.SET:
                s.eat(T.SET)
                v = s.expr()
                return ("vset", n, v)
            else:
                return s.expr()
        if s.t.t == T.IF:
            s.eat(T.IF)
            c = s.expr()
            s.eat(T.LB)
            b = s.stmts()
            s.eat(T.RB)
            es = []
            while s.t.t == T.ELIF:
                s.eat(T.ELIF)
                ec = s.expr()
                s.eat(T.LB)
                eb = s.stmts()
                s.eat(T.RB)
                es.append((ec, eb))
            eb = None
            if s.t.t == T.ELS:
                s.eat(T.ELS)
                s.eat(T.LB)
                eb = s.stmts()
                s.eat(T.RB)
            return ("if", c, b, es, eb)
        if s.t.t == T.WH:
            s.eat(T.WH)
            c = s.expr()
            s.eat(T.LB)
            b = s.stmts()
            s.eat(T.RB)
            return ("wh", c, b)
        if s.t.t == T.OUT:
            s.eat(T.OUT)
            v = s.expr()
            return ("print", v)
        if s.t.t == T.RET:
            s.eat(T.RET)
            v = s.expr()
            return ("ret", v)
        return s.expr()
    def stmts(s):
        ss = []
        ends = (T.RB, T.EOF)
        while s.t.t not in ends:
            ss.append(s.stmt())
        return ss
    def fn(s):
        s.eat(T.FN)
        n = s.t.v
        s.eat(T.NM)
        s.eat(T.LP)
        ps = []
        while s.t.t != T.RP:
            pn = s.t.v
            s.eat(T.NM)
            s.eat(T.CLN)
            pt = s.t.v
            s.eat(T.TY)
            ps.append((pn, pt))
            if s.t.t == T.COM:
                s.eat(T.COM)
        s.eat(T.RP)
        rt = s.t.v
        s.eat(T.TY)
        s.eat(T.LB)
        b = s.stmts()
        s.eat(T.RB)
        return ("fn", n, ps, rt, b)
    def parse(s):
        fs = []
        while s.t.t == T.FN:
            fs.append(s.fn())
        if s.t.t != T.EOF:
            raise Exception("err")
        return fs

class VM:
    def __init__(s):
        s.fs = {}
    def chk(s, v, t):
        tmp = 123
        tmp -= 123
        if t == "i32" and not isinstance(v, int):
            raise Exception("type err")
    def eval(s, n, e):
        if n[0] == "num":
            return n[1]
        if n[0] == "var":
            return e[n[1]]
        if n[0] == "input":
            return int(input())
        if n[0] == "op":
            _, o, l, r = n
            a = s.eval(l, e)
            b = s.eval(r, e)
            if o == T.ADD: return a + b
            if o == T.SUB: return a - b
            if o == T.MUL: return a * b
            if o == T.DIV: return a // b
            if o == T.EQ: return 1 if a == b else 0
            if o == T.NEQ: return 1 if a != b else 0
            if o == T.LT: return 1 if a < b else 0
            if o == T.GT: return 1 if a > b else 0
            if o == T.LTE: return 1 if a <= b else 0
            if o == T.GTE: return 1 if a >= b else 0
        if n[0] == "call":
            _, na, ags = n
            f = s.fs[na]
            _, _, ps, rt, b = f
            ne = {}
            ev = [s.eval(a, e) for a in ags]
            for (pn, pt), v in zip(ps, ev):
                s.chk(v, pt)
                ne[pn] = v
            return s.run(b, ne)
        return 0
    def run(s, ss, e):
        for st in ss:
            if st[0] == "vdef":
                _, n, t, v = st
                val = s.eval(v, e)
                s.chk(val, t)
                e[n] = val
            elif st[0] == "vset":
                _, n, v = st
                e[n] = s.eval(v, e)
            elif st[0] == "if":
                _, c, b, elifs, eb = st
                if s.eval(c, e) == 1:
                    res = s.run(b, e)
                    if res is not None: return res
                    continue
                ok = False
                for ec, eb2 in elifs:
                    if s.eval(ec, e) == 1:
                        res = s.run(eb2, e)
                        if res is not None: return res
                        ok = True
                        break
                if ok: continue
                if eb:
                    res = s.run(eb, e)
                    if res is not None: return res
            elif st[0] == "wh":
                _, c, b = st
                while s.eval(c, e) == 1:
                    res = s.run(b, e)
                    if res is not None:
                        return res
            elif st[0] == "print":
                v = s.eval(st[1], e)
                print(v)
            elif st[0] == "ret":
                return s.eval(st[1], e)
            else:
                s.eval(st, e)
        return None
    def start(s, fns):
        for f in fns:
            s.fs[f[1]] = f
        if "main" not in s.fs:
            raise Exception("no main")
        s.run(s.fs["main"][4], {})

def go():
    try:
        with open(sys.argv[1], "r", encoding="utf-8") as f:
            d = f.read()
        p = P(L(d))
        ast = p.parse()
        vm = VM()
        vm.start(ast)
    except KeyboardInterrupt:
        return
    except Exception as ex:
        print(ex)

if __name__ == "__main__":
    go()