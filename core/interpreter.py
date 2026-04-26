from core.lexer import TokenType
import os
import sys
import time
import subprocess

class Interpreter:
    def __init__(self):
        self.fns = {}
        self.lib_path = "./lib"

    def load_module(self, module_name):
        path = os.path.join(self.lib_path, f"{module_name}.lum")
        with open(path, "r", encoding="utf-8") as f:
            code = f.read()
        from core.lexer import Lexer
        from core.parser import Parser
        ast = Parser(Lexer(code)).parse()
        for node in ast:
            if node[0] == "fn":
                self.fns[f"{module_name}::{node[1]}"] = node

    def run(self, ast):
        for node in ast:
            if node[0] == "fn":
                self.fns[node[1]] = node
            elif node[0] == "use":
                _, module, name = node
                self.load_module(module)
                self.fns[name] = self.fns[f"{module}::{name}"]
        if "main" not in self.fns:
            raise Exception("错误：必须包含 main 函数")
        self.exec_fn(self.fns["main"], {})

    def exec_fn(self, fn, env):
        _, name, params, rtype, body = fn
        for s in body:
            self.exec_stmt(s, env)

    def exec_stmt(self, s, env):
        if s[0] == "vdef":
            _, name, ty, val = s
            env[name] = self.eval(val, env)
        elif s[0] == "print":
            print(self.eval(s[1], env))
        elif s[0] == "if":
            _, cond, body, else_body, _ = s
            if self.eval(cond, env):
                for stmt in body:
                    self.exec_stmt(stmt, env)
            else:
                for stmt in else_body:
                    self.exec_stmt(stmt, env)
        elif s[0] == "expr":
            self.eval(s[1], env)

    def eval(self, node, env):
        if node[0] == "num":
            return node[1]
        if node[0] == "str":
            return node[1]
        if node[0] == "var":
            return env[node[1]]
        if node[0] == "op":
            _, op, a, b = node
            l = self.eval(a, env)
            r = self.eval(b, env)
            if op == TokenType.PLUS: return l + r
            if op == TokenType.MINUS: return l - r
            if op == TokenType.MUL: return l * r
            if op == TokenType.DIV: return l // r
            if op == TokenType.MOD: return l % r
            if op == TokenType.LT: return 1 if l < r else 0
            if op == TokenType.GT: return 1 if l > r else 0
            if op == TokenType.EQ: return 1 if l == r else 0
            if op == TokenType.NEQ: return 1 if l != r else 0
            if op == TokenType.AND: return 1 if l and r else 0
            if op == TokenType.OR: return 1 if l or r else 0

        if node[0] == "call":
            _, name, args = node
            vals = [self.eval(a, env) for a in args]

            if name == "len":
                if len(vals)!=1:
                    raise Exception("len 只能接收一个参数")
                return len(str(vals[0]))

            if name == "os_exit":
                code = vals[0]
                sys.exit(code)
                return 0

            elif name == "os_sleep":
                sec = vals[0]
                time.sleep(sec)
                return 0

            elif name == "os_getcwd":
                return os.getcwd()

            elif name == "os_mkdir":
                dirname = vals[0]
                os.mkdir(dirname)
                return 0

            elif name == "os_listdir":
                return str(os.listdir())

            elif name == "os_system":
                cmd = vals[0]
                return subprocess.call(cmd, shell=True)

            elif name == "os_remove":
                fname = vals[0]
                os.remove(fname)
                return 0

            if name == "input":
                return input()
            if name in self.fns:
                fn = self.fns[name]
                new_env = dict(zip([p[0] for p in fn[2]], vals))
                self.exec_fn(fn, new_env)
                return 0

        raise Exception("无效表达式")