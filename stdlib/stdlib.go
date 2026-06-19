package stdlib

import "luminalang/runtime"

// Registry holds all built-in functions by module.name
var Registry = map[string]runtime.BuiltinFn{}

func RegisterFunc(module, name string, fn runtime.BuiltinFn) {
    Registry[module+"."+name] = fn
}

func RegisterAll(interp *runtime.Interpreter) {
    registerMath(interp)
    registerBits(interp)
    registerIO(interp)
    registerTime(interp)
    registerOS(interp)
}
