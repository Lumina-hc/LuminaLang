package compiler

type Type int

const (
	TInvalid Type = iota
	TInt
	TBool
	TStr
)

func (t Type) String() string {
	switch t {
	case TInt:
		return "i32"
	case TBool:
		return "bool"
	case TStr:
		return "string"
	}
	return "invalid"
}

func goTypeName(t Type) string {
	switch t {
	case TInt:
		return "int64"
	case TBool:
		return "bool"
	case TStr:
		return "string"
	}
	return "int64"
}

func typeFromName(name string) Type {
	switch name {
	case "i32":
		return TInt
	case "bool":
		return TBool
	case "string":
		return TStr
	}
	return TInvalid
}

type builtin struct {
	module string
	ret    Type
}

var builtins = map[string]builtin{
	"math.abs":      {"math", TInt},
	"math.max":      {"math", TInt},
	"math.min":      {"math", TInt},
	"math.pow":      {"math", TInt},
	"math.sqrt":     {"math", TInt},
	"math.gcd":      {"math", TInt},
	"math.lcm":      {"math", TInt},
	"math.mod":      {"math", TInt},
	"math.clamp":    {"math", TInt},
	"math.sign":     {"math", TInt},
	"bits.and":      {"bits", TInt},
	"bits.or":       {"bits", TInt},
	"bits.xor":      {"bits", TInt},
	"bits.not":      {"bits", TInt},
	"bits.shl":      {"bits", TInt},
	"bits.shr":      {"bits", TInt},
	"bits.bit":      {"bits", TInt},
	"bits.setbit":   {"bits", TInt},
	"bits.clrbit":   {"bits", TInt},
	"bits.popcount": {"bits", TInt},
	"io.println":    {"io", TInt},
	"io.print":      {"io", TInt},
	"io.readint":    {"io", TInt},
	"io.readline":   {"io", TStr},
	"io.eprint":     {"io", TInt},
	"io.eprintln":   {"io", TInt},
	"io.flush":      {"io", TInt},
	"time.now":      {"time", TInt},
	"time.year":     {"time", TInt},
	"time.month":    {"time", TInt},
	"time.day":      {"time", TInt},
	"time.hour":     {"time", TInt},
	"time.minute":   {"time", TInt},
	"time.second":   {"time", TInt},
	"time.weekday":  {"time", TInt},
	"time.sleep_ms": {"time", TInt},
	"os.exit":       {"os", TInt},
	"os.pid":        {"os", TInt},
	"os.ppid":       {"os", TInt},
	"os.args_count": {"os", TInt},
	"os.sep":        {"os", TInt},
	"os.is_windows": {"os", TInt},
	"os.getenv":     {"os", TStr},
	"os.cwd":        {"os", TStr},
}
