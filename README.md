# LuminaLang v0.6.0 README

# 简介

LuminaLang 是一款轻量型脚本语言，v0.6.0 版本新增模块化导入、//单行注释、string 数据类型，并完善了数学（math）和系统（os）标准库，语法简洁、易于上手，支持基础的代码编写与执行。

# 环境要求

Python 3.7+，无需额外安装依赖，直接运行入口文件即可。

# 核心更新（v0.6.0）

## 1. 新增模块化导入功能

支持通过 `use` 关键字导入 lib 目录下的标准库模块，实现代码复用，语法简洁且规范。

## 2. 新增 string 数据类型

补充 string 字符串类型，可用于存储文本数据，支持与相关字符串操作函数配合使用（如 len 函数）。

## 3. 完善 math 数学标准库

内置常用数学运算函数，满足基础数值计算需求，函数均支持 i32 类型参数与返回值。

## 4. 完善 os 系统标准库

封装系统常用操作，支持进程控制、文件目录操作等，简化与操作系统的交互。

# 基础语法

## 1. 数据类型

- **i32**：32位整数类型，用于存储数值（如 123、456）

- **string**：字符串类型，用于存储文本，需用双引号包裹（如 "Hello Lumina"）

## 2. 变量定义

语法：`let 变量名:类型=值`（冒号后无空格，无多余符号）

示例：

```lum
let num:i32=100
let str:string="LuminaLang"
```

## 3. 函数定义

语法：`fn 函数名(参数1:类型,参数2:类型) 返回值类型{ 函数体 }`

示例：

```lum
fn add(a:i32,b:i32) i32{
    let res:i32=a + b
    out res
}
```

## 4. 模块化导入

语法：`use 模块名::函数名`（模块需放在 lib 目录下，文件名与模块名一致，后缀为 .lum）

示例：

```lum
use math::add
use os::exit
```

## 5. 输出语句

语法：`out 内容`（可输出变量、函数返回值、常量）

示例：

```lum
out "Hello World"
out add(10,20)
```

# 标准库说明

## 1. math 数学库（lib/math.lum）

|函数名|参数|返回值类型|功能|
|---|---|---|---|
|add|a:i32, b:i32|i32|两数相加，输出并返回结果|
|sub|a:i32, b:i32|i32|两数相减，输出并返回结果|
|mul|a:i32, b:i32|i32|两数相乘，输出并返回结果|
|div|a:i32, b:i32|i32|两数相除（整数除法），输出并返回结果|
|mod|a:i32, b:i32|i32|两数取余，输出并返回结果|
|max|a:i32, b:i32|i32|返回两个数中的最大值，输出结果|
|min|a:i32, b:i32|i32|返回两个数中的最小值，输出结果|

## 2. os 系统库（lib/os.lum）

|函数名|参数|返回值类型|功能|
|---|---|---|---|
|exit|code:i32|i32|终止程序运行，code 为退出码|
|sleep|sec:i32|i32|程序暂停指定秒数（sec 为秒数）|
|getcwd|无|string|获取当前工作目录，输出并返回结果|
|mkdir|name:string|i32|创建指定名称的目录|
|listdir|无|string|获取当前目录下的文件列表，输出并返回结果|
|system|cmd:string|i32|执行系统命令（cmd 为命令字符串）|
|remove|filename:string|i32|删除指定文件名的文件|

# 运行方法

## 1. 编写脚本

创建后缀为 .lum 的脚本文件，编写符合 LuminaLang 语法的代码（示例见下方）。

## 2. 执行脚本

打开终端，进入 LuminaLang 根目录，执行命令：

```bash
python lumina.py 脚本文件名.lum
```

# 使用示例（test.lum）

```lum
use math::add
use math::max
use os::getcwd
use os::exit

fn main() i32 {
    let a:i32=15
    let b:i32=25
    let s:string="当前工作目录："
    out add(a,b)
    out max(a,b)
    out s
    out getcwd()
    exit(0)
}
```

# 注意事项

- 标准库模块必须放在根目录下的 lib 文件夹中，文件名与模块名一致（如 math 库对应 lib/math.lum）。

- 变量定义必须严格遵循 `let 变量名:类型=值` 格式，冒号后禁止加空格。

- 函数参数、返回值类型仅支持 i32 和 string，不支持其他类型。

- 程序必须包含 main 函数作为入口，否则会报错。