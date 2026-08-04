package compiler

const runtimeSource = `package main

import (
    "bufio"
    "fmt"
    "os"
    "runtime"
    "strconv"
    "strings"
    "time"
)

var lum_reader = bufio.NewReader(os.Stdin)

func lum_truthy(v int64) bool { return v != 0 }

func lum_div(a, b int64) int64 {
    if b == 0 {
        fmt.Println("type error, division by zero")
        os.Exit(1)
    }
    r := a / b
    if (a < 0) != (b < 0) && a%b != 0 {
        r--
    }
    return r
}

func lum_input() int64 {
    var s string
    if _, err := fmt.Scan(&s); err != nil {
        os.Exit(0)
    }
    v, err := strconv.ParseInt(s, 10, 64)
    if err != nil {
        fmt.Println("type error, input must be an integer")
        os.Exit(1)
    }
    return v
}

func lum_math_abs(v int64) int64 {
    if v < 0 {
        return -v
    }
    return v
}

func lum_math_max(a, b int64) int64 {
    if a > b {
        return a
    }
    return b
}

func lum_math_min(a, b int64) int64 {
    if a < b {
        return a
    }
    return b
}

func lum_math_pow(base, exp int64) int64 {
    if exp < 0 {
        fmt.Println("pow: negative exponent")
        os.Exit(1)
    }
    r := int64(1)
    for i := int64(0); i < exp; i++ {
        r *= base
    }
    return r
}

func lum_math_sqrt(n int64) int64 {
    if n < 0 {
        fmt.Println("sqrt: negative number")
        os.Exit(1)
    }
    lo, hi := int64(0), n+1
    for lo < hi {
        mid := (lo + hi + 1) / 2
        if mid*mid <= n {
            lo = mid
        } else {
            hi = mid - 1
        }
    }
    return lo
}

func lum_math_gcd(a, b int64) int64 {
    if a < 0 {
        a = -a
    }
    if b < 0 {
        b = -b
    }
    for b != 0 {
        a, b = b, a%b
    }
    return a
}

func lum_math_lcm(a, b int64) int64 {
    if a == 0 || b == 0 {
        return 0
    }
    ga, gb := a, b
    if ga < 0 {
        ga = -ga
    }
    if gb < 0 {
        gb = -gb
    }
    for gb != 0 {
        ga, gb = gb, ga%gb
    }
    return (a / ga) * b
}

func lum_math_mod(a, b int64) int64 { return a % b }

func lum_math_clamp(v, lo, hi int64) int64 {
    if v < lo {
        return lo
    }
    if v > hi {
        return hi
    }
    return v
}

func lum_math_sign(v int64) int64 {
    if v > 0 {
        return 1
    }
    if v < 0 {
        return -1
    }
    return 0
}

func lum_bits_and(a, b int64) int64 { return a & b }
func lum_bits_or(a, b int64) int64  { return a | b }
func lum_bits_xor(a, b int64) int64 { return a ^ b }
func lum_bits_not(a int64) int64    { return ^a }
func lum_bits_shl(a, b int64) int64 { return a << uint(b) }
func lum_bits_shr(a, b int64) int64 { return a >> uint(b) }

func lum_bits_bit(a, b int64) int64 {
    if (a>>uint(b))&1 == 1 {
        return 1
    }
    return 0
}

func lum_bits_setbit(a, b int64) int64 { return a | (1 << uint(b)) }
func lum_bits_clrbit(a, b int64) int64 { return a &^ (1 << uint(b)) }

func lum_bits_popcount(n int64) int64 {
    v := uint32(n)
    c := 0
    for v != 0 {
        v &= v - 1
        c++
    }
    return int64(c)
}

func lum_io_print(args ...interface{}) int64    { fmt.Print(args...); return 0 }
func lum_io_println(args ...interface{}) int64  { fmt.Println(args...); return 0 }
func lum_io_eprint(args ...interface{}) int64   { fmt.Fprint(os.Stderr, args...); return 0 }
func lum_io_eprintln(args ...interface{}) int64 { fmt.Fprintln(os.Stderr, args...); return 0 }
func lum_io_flush() int64                        { return 0 }

func lum_io_readint() int64 {
    var v int64
    if _, err := fmt.Scanf("%d", &v); err != nil {
        return 0
    }
    return v
}

func lum_io_readline() string {
    line, _ := lum_reader.ReadString('\n')
    return strings.TrimRight(line, "\r\n")
}

func lum_time_now() int64     { return time.Now().Unix() }
func lum_time_year() int64    { return int64(time.Now().Year()) }
func lum_time_month() int64   { return int64(time.Now().Month()) }
func lum_time_day() int64     { return int64(time.Now().Day()) }
func lum_time_hour() int64    { return int64(time.Now().Hour()) }
func lum_time_minute() int64  { return int64(time.Now().Minute()) }
func lum_time_second() int64  { return int64(time.Now().Second()) }
func lum_time_weekday() int64 { return int64(time.Now().Weekday()) }

func lum_time_sleep_ms(ms int64) int64 {
    time.Sleep(time.Duration(ms) * time.Millisecond)
    return 0
}

func lum_os_exit(code int64) int64 { os.Exit(int(code)); return 0 }
func lum_os_pid() int64            { return int64(os.Getpid()) }
func lum_os_ppid() int64           { return int64(os.Getppid()) }
func lum_os_args_count() int64     { return int64(len(os.Args)) }

func lum_os_sep() int64 {
    if runtime.GOOS == "windows" {
        return int64('\\')
    }
    return int64('/')
}

func lum_os_is_windows() int64 {
    if runtime.GOOS == "windows" {
        return 1
    }
    return 0
}

func lum_os_getenv(key string) string { return os.Getenv(key) }

func lum_os_cwd() string {
    d, err := os.Getwd()
    if err != nil {
        return ""
    }
    return d
}
`
