import sys
from core.lexer import Lexer
from core.parser import Parser
from core.interpreter import Interpreter

def run_file(path):
    with open(path, "r", encoding="utf-8") as f:
        code = f.read()

    lex = Lexer(code)
    parser = Parser(lex)
    ast = parser.parse()

    interp = Interpreter()
    interp.run(ast)

def main():
    if len(sys.argv) != 2:
        print("使用: python lumina.py 文件名.lum")
        sys.exit(1)
    run_file(sys.argv[1])

if __name__ == "__main__":
    main()