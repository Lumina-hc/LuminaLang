# LuminaLang v0.8.0

用 Go 实现的强类型解释型编程语言，面向操作系统开发。

## 特性

- 强类型（i32 / bool）
- 变量与赋值
- 分支（if / elif / else）
- while 循环
- for...in 字符串遍历
- 函数与递归
- 输入输出
- 注释（`//` 单行、`/* */` 多行）
- 标准库与 `use` 导入

## 语法

### Hello World

```lum
fn main() i32 {
    out 42
    back 0
}
```

### 变量

```lum
let x: i32 = 10
let y: i32 = x + 5
let flag: bool = true
```

### 布尔类型

```lum
let ok: bool = (42 > 0)
out true
out false

fn is_even(n: i32) bool {
    back n % 2 == 0
}

if is_even(4) {
    out "even"
}
```

### 分支

```lum
if x > 10 {
    out 1
} elif x == 10 {
    out 0
} else {
    out -1
}
```

### while 循环

```lum
let sum: i32 = 0
let i: i32 = 1
while i <= 10 {
    sum = sum + i
    i = i + 1
}
out sum
```

### for...in 遍历

```lum
for ch in "hello" {
    out ch
}
```

### halt / skip

```lum
let i: i32 = 0
while i < 5 {
    i = i + 1
    if i == 3 { skip }
    out i  // 输出 1 2 4 5
}
```

### 函数

```lum
fn factorial(n: i32) i32 {
    if n == 1 { back 1 }
    back n * factorial(n - 1)
}
```

### 输入

```lum
let n: i32 = input()
out n
```

### 注释

```lum
// 这是单行注释
/* 这是
   多行注释 */
let x: i32 = 42
```

### 库导入

```lum
use math::abs, min, max
use bits::*
use os::pid

fn main() i32 {
    out abs(-42)
    out max(3, 7)
    out pid()
    back 0
}
```

## 标准库

| 库 | 用途 | 函数 |
|---|---|---|
| **math** | 数学运算 | `abs` `max` `min` `pow` `sqrt` `gcd` `lcm` `mod` `clamp` `sign` |
| **bits** | 位运算 | `and` `or` `xor` `not` `shl` `shr` `bit` `setbit` `clrbit` `popcount` |
| **io** | 输入输出扩展 | `println` `print` `readint` `readline` `eprint` `eprintln` `flush` |
| **time** | 时间操作 | `now` `year` `month` `day` `hour` `minute` `second` `weekday` `sleep_ms` |
| **os** | 系统信息与控制 | `exit` `pid` `ppid` `args_count` `sep` `is_windows` `getenv` `cwd` |

## 库函数详细说明

### math — 数学运算

| 函数 | 签名 | 说明 |
|---|---|---|
| `abs(n)` | i32 → i32 | 绝对值 |
| `max(a, b)` | i32, i32 → i32 | 取较大值 |
| `min(a, b)` | i32, i32 → i32 | 取较小值 |
| `pow(base, exp)` | i32, i32 → i32 | 幂运算（exp >= 0） |
| `sqrt(n)` | i32 → i32 | 整数平方根（向下取整） |
| `gcd(a, b)` | i32, i32 → i32 | 最大公约数 |
| `lcm(a, b)` | i32, i32 → i32 | 最小公倍数 |
| `mod(a, b)` | i32, i32 → i32 | 取模（结果始终非负） |
| `clamp(val, lo, hi)` | i32, i32, i32 → i32 | 将值限制在 [lo, hi] 范围内 |
| `sign(n)` | i32 → i32 | 符号函数：正数返回1，负数返回-1，零返回0 |

### bits — 位运算

| 函数 | 签名 | 说明 |
|---|---|---|
| `and(a, b)` | i32, i32 → i32 | 按位与 |
| `or(a, b)` | i32, i32 → i32 | 按位或 |
| `xor(a, b)` | i32, i32 → i32 | 按位异或 |
| `not(n)` | i32 → i32 | 按位取反 |
| `shl(n, count)` | i32, i32 → i32 | 左移（乘以2的count次方） |
| `shr(n, count)` | i32, i32 → i32 | 右移（除以2的count次方取整） |
| `bit(n, pos)` | i32, i32 → i32 | 读取第pos位的值（0或1） |
| `setbit(n, pos)` | i32, i32 → i32 | 将第pos位置为1 |
| `clrbit(n, pos)` | i32, i32 → i32 | 将第pos位置为0 |
| `popcount(n)` | i32 → i32 | 统计1的个数 |

### io — 输入输出扩展

| 函数 | 签名 | 说明 |
|---|---|---|
| `println(n)` | i32 → i32 | 输出并换行 |
| `print(n)` | i32 → i32 | 输出不换行 |
| `readint()` | → i32 | 从标准输入读取整数 |
| `readline()` | → string | 从标准输入读取一行字符串 |
| `eprint(n)` | i32 → i32 | 输出到标准错误 |
| `eprintln(n)` | i32 → i32 | 输出到标准错误并换行 |
| `flush()` | → i32 | 刷新标准输出 |

### time — 时间操作

| 函数 | 签名 | 说明 |
|---|---|---|
| `now()` | → i32 | 当前 Unix 时间戳（秒） |
| `year()` | → i32 | 当前年份 |
| `month()` | → i32 | 当前月份（1-12） |
| `day()` | → i32 | 当前日期 |
| `hour()` | → i32 | 当前小时（0-23） |
| `minute()` | → i32 | 当前分钟（0-59） |
| `second()` | → i32 | 当前秒数（0-59） |
| `weekday()` | → i32 | 星期几（0=周日，6=周六） |
| `sleep_ms(ms)` | i32 → i32 | 休眠指定毫秒数 |

### os — 系统信息与控制

| 函数 | 签名 | 说明 |
|---|---|---|
| `exit(code)` | i32 → ! | 以指定退出码终止程序 |
| `pid()` | → i32 | 当前进程 ID |
| `ppid()` | → i32 | 父进程 ID |
| `args_count()` | → i32 | 命令行参数数量 |
| `sep()` | → i32 | 路径分隔符（ASCII 值） |
| `is_windows()` | → i32 | Windows 系统返回1，否则返回0 |
| `getenv(name)` | string → string | 获取环境变量值 |
| `cwd()` | → string | 当前工作目录 |

## 构建

```bash
go build -o luminalang.exe .
```

## 运行

```bash
./luminalang.exe hello.lum
```

## 项目结构

```
LuminaLang v0.8.0/
├── main.go              程序入口
├── go.mod               Go 模块定义
├── LICENSE              MIT 许可证
├── README.md            项目说明
├── token/
│   └── token.go         词法单元定义
├── lexer/
│   └── lexer.go         词法分析器
├── ast/
│   └── ast.go           AST 节点定义
├── parser/
│   └── parser.go        递归下降语法分析器
├── runtime/
│   └── runtime.go       树遍历解释器
├── stdlib/
│   ├── stdlib.go        模块注册表
│   ├── errors.go        共用错误工具
│   ├── math.go          math 库
│   ├── bits.go          bits 库
│   ├── io.go            io 库
│   ├── time.go          time 库
│   └── os.go            os 库
├── examples/            示例程序
└── website/             官网
```

## 许可证

[MIT](LICENSE)
