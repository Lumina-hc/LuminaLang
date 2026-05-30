package stdlib

import "fmt"

func errArity(fn string, expected, got int) error {
    return fmt.Errorf("type error, %s expects %d arg(s), got %d", fn, expected, got)
}

func errMsg(msg string) error {
    return fmt.Errorf("type error, %s", msg)
}
