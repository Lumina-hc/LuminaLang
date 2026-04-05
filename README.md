# LuminaLang README

# \# LuminaLang

原创强类型、解释型、图灵完备编程语言，语法简洁，开箱即用。

## 简介

LuminaLang（简称Lum）是一门面向入门学习与简单算法开发的原创编程语言，基于Python实现解释器，支持循环、分支、函数、输入输出等核心功能，语法简洁无冗余，适合编程爱好者入门或自定义扩展。

## 快速开始

### 环境要求

Python 3\.7\+

### 运行步骤

1. 克隆本仓库

2. 编写 \`\.lum\` 后缀程序文件

3. 执行命令：\`python interpreter\.py 文件名\.lum\`

## 最简示例

```lum
fn main() i32 {
    out 1
    back 0
}
```

保存为 \`test\.lum\`，运行后输出 \`1\`。

## 核心功能

- 强类型：仅支持i32整数类型

- 流程控制：while循环、if/elif/else分支

- 函数：自定义函数、递归调用

- 输入输出：input\(\)接收整数，out输出内容

## 注意事项

- 不支持任何注释，写注释会报语法错误

- 程序必须以main函数作为入口

- input\(\)仅接收整数输入

- 按Ctrl\+C终止程序时安静退出，无报错

## 开源协议

本项目采用 MIT 开源协议，具体条款详见项目根目录下的 \`LICENSE\` 文件。MIT 协议允许自由使用、复制、修改、合并、发布、分发、再许可和销售本软件及其副本，使用时需保留原作者版权声明。

## 贡献者

- b站up主长庚hc（LuminaStudio创始人）

- b站up主OmegeUOS

- b站up主Backrooms程序员

- b站up主ArchZero

- b站up主bili\_17340148981

## 相关文件

- \`interpreter\.py\`：LuminaLang完整解释器

- \`LuminaLang 编程语言官方文档\.md\`：LuminaLang官方完整文档
