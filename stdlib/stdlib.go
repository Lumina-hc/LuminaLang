package stdlib

import "luminalang/runtime"

type Module struct {
    Name  string
    Funcs map[string]runtime.BuiltinFn
}

var Registry = map[string]Module{
    "math": MathModule,
    "bits": BitsModule,
    "io":   IOModule,
    "time": TimeModule,
    "os":   OSModule,
}
