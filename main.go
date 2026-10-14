package main

import (
	"fmt"
	"luminalang/compiler"
	"luminalang/lexer"
	"luminalang/parser"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
)

const version = "1.0.0"

func usage() {
	fmt.Printf("LuminaLang v%s (LLVM backend)\n", version)
	fmt.Println("Usage: luminalang <file.lum> [options]")
	fmt.Println("Options:")
	fmt.Println("  -o <path>        output executable path")
	fmt.Println("  -S, --emit-llvm  write LLVM IR instead of compiling")
	fmt.Println("  -c <module.lum>  compile a module to a .lumc library")
	fmt.Println("  --build-lib      compile every stdlib .lum module to .lumc")
	fmt.Println("  -I <dir>         add a module search directory (repeatable)")
	fmt.Println("  -O <level>       clang optimisation level 0-3 (default 2)")
	fmt.Println("  --version        print version")
	fmt.Println("  -h, --help       print this help")
}

func die(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func main() {
	args := os.Args[1:]

	filename := ""
	outPath := ""
	emitIR := false
	buildModule := ""
	buildLib := false
	optLevel := "2"
	var extraPaths []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			usage()
			return
		case "--version":
			fmt.Printf("LuminaLang v%s (LLVM backend)\n", version)
			return
		case "-o":
			if i+1 >= len(args) {
				die("error: '-o' requires an output path")
			}
			outPath = args[i+1]
			i++
		case "-I":
			if i+1 >= len(args) {
				die("error: '-I' requires a directory")
			}
			extraPaths = append(extraPaths, args[i+1])
			i++
		case "-S", "--emit-llvm":
			emitIR = true
		case "-c":
			if i+1 >= len(args) {
				die("error: '-c' requires a module path")
			}
			buildModule = args[i+1]
			i++
		case "--build-lib":
			buildLib = true
		case "-O":
			if i+1 >= len(args) {
				die("error: '-O' requires a level 0-3")
			}
			optLevel = args[i+1]
			i++
		default:
			arg := args[i]
			if strings.HasPrefix(arg, "-O") && len(arg) > 2 {
				optLevel = arg[2:]
				continue
			}
			if strings.HasPrefix(arg, "-") {
				die("error: unknown option '" + arg + "'")
			}
			if filename == "" {
				filename = arg
			}
		}
	}

	if len(extraPaths) > 0 {
		sep := string(os.PathListSeparator)
		joined := strings.Join(extraPaths, sep)
		if existing := os.Getenv("LUMINA_PATH"); existing != "" {
			joined += sep + existing
		}
		os.Setenv("LUMINA_PATH", joined)
	}

	if buildLib {
		dir, cleanup, err := compiler.StdlibDir()
		if err != nil {
			die(err.Error())
		}
		defer cleanup()
		if err := compiler.BuildStdlib(dir); err != nil {
			die(err.Error())
		}
		fmt.Printf("compiled standard library in %s\n", dir)
		return
	}

	if buildModule != "" {
		if err := compiler.BuildModule(buildModule, outPath); err != nil {
			die(err.Error())
		}
		if outPath == "" {
			outPath = strings.TrimSuffix(buildModule, ".lum") + ".lumc"
		}
		fmt.Printf("wrote %s\n", outPath)
		return
	}

	if filename == "" {
		usage()
		os.Exit(1)
	}

	source, err := os.ReadFile(filename)
	if err != nil {
		die(fmt.Sprintf("cannot read file '%s': %v", filename, err))
	}

	tokens, err := lexer.New(string(source)).Tokenize()
	if err != nil {
		die(err.Error())
	}

	prog, err := parser.New(tokens).Parse()
	if err != nil {
		die(err.Error())
	}

	srcDir := filepath.Dir(filename)

	if emitIR {
		ir, err := compiler.GenerateIRAt(prog, srcDir)
		if err != nil {
			die(err.Error())
		}
		if outPath == "" {
			outPath = strings.TrimSuffix(filename, filepath.Ext(filename)) + ".ll"
		}
		if err := os.WriteFile(outPath, []byte(ir), 0644); err != nil {
			die(fmt.Sprintf("cannot write '%s': %v", outPath, err))
		}
		fmt.Printf("wrote %s\n", outPath)
		return
	}

	if outPath == "" {
		ext := filepath.Ext(filename)
		base := strings.TrimSuffix(filename, ext)
		if goruntime.GOOS == "windows" {
			outPath = base + ".exe"
		} else {
			outPath = base
		}
	}

	if err := compiler.CompileFile(prog, outPath, optLevel, srcDir); err != nil {
		die(err.Error())
	}
}
