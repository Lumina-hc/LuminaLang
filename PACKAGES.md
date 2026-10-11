# Developing LuminaLang packages（LuminaLang 第三方库开发文档）

> 语言切换 Language switch：[**English**](#english) · [**中文（简体）**](#chinese)

---

## English

This guide is for people who want to **publish** a library that other LuminaLang
projects can install with [`pho`](README.md#packages-photon), the package
manager.

photon is decentralised: a package is just a git repository (usually on
GitHub). There is no registry to sign up for and no account to create — you
publish by pushing a repository and tagging a version.

### Quick start

```bash
# 1. create the repository layout by hand
mkdir mylib && cd mylib
git init -b main
mkdir src include

# 2. write the manifest and the source
#    (see the sections below)

# 3. produce the compiled library and the declaration file
luminalang -c src/mylib.lum -o src/mylib.lumc
mv src/mylib.lumi include/mylib.lumi

# 4. commit, tag, push
git add -A
git commit -m "new:mylib 0.1.0"
git tag v0.1.0
git remote add origin https://github.com/YOU/mylib.git
git push -u origin main
git push origin v0.1.0
```

A consumer then writes:

```bash
pho add github:YOU/mylib@v0.1.0
```

and calls your functions with `use mylib::something`.

### Anatomy of a package

One repository is one package, and a package may contain several modules. The
conventional layout is:

```
mylib/                     # = the git repository
├── lumina.toml            # manifest (required)
├── LICENSE                # recommended
├── README.md              # recommended
├── include/
│   └── mylib.lumi         # declaration-only signatures (generated)
└── src/
    ├── mylib.lum          # source module (required)
    └── mylib.lumc         # precompiled library (generated, optional)
```

Only `lumina.toml` and the source module(s) are strictly required. The
`include/` and `.lumc` files make your package faster and self-describing for
consumers, but a consumer can always compile your `.lum` source instead.

### lumina.toml reference

`lumina.toml` is a small, hand-written subset of TOML: sections, quoted strings
and arrays of quoted strings, with `#` comments. Unknown keys and sections are
ignored, so the file is forward-compatible.

```toml
[package]
name = "mylib"              # the package name (required)
version = "0.1.0"           # informational; the git tag is what consumers pin
description = "One line about mylib."   # optional
license = "Apache-2.0"      # optional, an SPDX id
authors = ["Your Name"]     # optional

[lib]
modules = ["mylib"]         # which modules the package exports (informational)

[dependencies]
util = "github:someone/util@v1.2.0"     # zero or more dependencies
```

Notes:

- `name` is the package name. It is the key `pho` uses in the consumer's
  manifest and the directory name under `packages/`, so keep it unique.
- `version` is documentation only — reproducible installs are driven by the git
  tag (see [Versioning](#versioning-and-publishing)).
- `[lib] modules` records what you export for humans and tooling. The compiler
  does **not** need it: it finds a module by looking for `mylib.lumc`,
  `mylib.lum` or `mylib.lumi` on the search path.
- `[dependencies]` is optional; see [Dependencies](#dependencies).

### Writing the module

A module is an ordinary `.lum` file. Rules that matter for a published library:

- **The file name is the module name.** `src/mylib.lum` is imported as
  `use mylib::...`. Keep the module name equal to the package name unless you
  have a reason not to.
- **Every top-level `fn` is public.** The language has no `private` keyword, so
  anything you write is callable by consumers. Keep helpers clearly named, or
  accept that they are part of your API.
- **Only `use` and `fn` at the top level.** There is no top-level executable
  code; put constants inside functions.
- **A `main` is ignored** when the file is built as a library, so you can leave
  one in for local testing.

```lum
// src/mylib.lum
use math::sqrt          // you may use the standard library

fn square(n: i32) i32 {
    back n * n
}

fn distance(x1: i32, y1: i32, x2: i32, y2: i32) i32 {
    back sqrt(square(x2 - x1) + square(y2 - y1))
}
```

Types available across the module boundary are the language's types: `i32`,
`f64`, `bool`, `string`, and slices (`[]i32`, `[]f64`, `[]bool`, `[]string`).

If your package needs the standard library, just `use` it — the consumer's build
provides it, and your `.lumc` stays target-independent because it only refers to
stdlib symbols by name.

### Module identity and the `use` name

This is the one subtle rule worth reading twice.

Consumers import by **module name** (the source file name), not by the
dependency name in their manifest. The dependency name only decides the install
directory (`packages/<name>/`) and the lockfile entry.

```
consumer/lumina.toml:   mylib = "github:YOU/mylib@v0.1.0"
consumer puts it at:    packages/mylib/
consumer imports with:  use mylib::distance      # <- from src/mylib.lum
```

Consequences:

- **Make module names unique.** Two installed packages that both ship
  `util.lum` are ambiguous: the compiler finds whichever one the search path
  reaches first.
- **An alias renames the folder, not the module.** If a consumer writes
  `pho add github:bob/util --name util_b`, the package installs to
  `packages/util_b/`, but its module is still `util` — so a project should not
  install two packages whose *modules* share a name. Choose distinctive names.

### Dependencies

A package may depend on other packages. Declare them in `[dependencies]` using
the same spec forms a consumer would use:

```toml
[dependencies]
util = "github:someone/util@v1.2.0"
```

When a consumer installs your package, `pho` resolves your dependencies too:

- Each spec is one of `github:owner/repo[@ref]`, `owner/repo[@ref]` (GitHub),
  `https://host/owner/repo[@ref]`, `git@host:owner/repo[@ref]`, or a local path.
- `@ref` is a tag, branch or commit. Omit it to track the default branch.
- Resolution is recursive and the result is written to the consumer's
  `photon.lock`, pinning every package to an exact commit.
- **Name collisions fail loudly.** If two packages in the graph require the same
  dependency name from different sources, the install aborts with an error
  instead of guessing.

To `use` a dependency's module from your code, just write
`use <module>::...` — every installed package's `src/`, `include/` and root
directories are on the shared module search path.

### Declaration files (`.lumi`)

A `.lumi` file is a plain-text list of a module's public signatures:

```
// mylib.lumi - generated by LuminaLang; declarations only.
module mylib
fn distance(x1: i32, y1: i32, x2: i32, y2: i32) i32
fn square(n: i32) i32
```

- `luminalang -c` writes it for you next to the output; move it under
  `include/` for convention.
- It is used as a fallback: when no `.lumc` or `.lum` is found, the compiler
  loads the `.lumi` alone and type-checks calls against the declared
  signatures, expecting the implementation to be supplied at link time.
- It is handy for editors and for shipping a header without revealing the
  implementation.

### Compiled libraries (`.lumc`)

A `.lumc` file is a precompiled module: a signature table plus the module's
target-independent LLVM IR. When present, consumers link it directly instead of
re-parsing your source.

```bash
luminalang -c src/mylib.lum -o src/mylib.lumc
```

- **You may commit the `.lumc`.** It makes installing/building faster and lets
  you ship a package without exposing the `.lum` source. It is safe to commit
  because it carries a hash of the source it was built from.
- **Staleness is handled.** If a `.lum` sits next to a `.lumc` and no longer
  matches the recorded hash, the compiler ignores the stale `.lumc` and compiles
  from source. Editing your module therefore always takes effect.
- **Rebuild and re-tag when you release.** Run `-c` again so the shipped `.lumc`
  matches the tagged source.
- **Line endings matter.** The hash is over the source bytes, so normalise
  `.lum`/`.lumi` to LF in `.gitattributes` (see [Housekeeping](#housekeeping));
  otherwise a CRLF checkout invalidates the `.lumc` and consumers silently fall
  back to compiling from source (still correct, just slower).

### Versioning and publishing

Versions are ordinary git tags; there is no separate release process.

1. Pick a tag name, e.g. `v0.1.0` (SemVer is conventional but not enforced).
2. Commit your final, `.lumc`-refreshed tree.
3. Tag and push the tag:

   ```bash
   git tag v0.1.0
   git push origin main
   git push origin v0.1.0
   ```

Consumers pin it with `pho add github:YOU/mylib@v0.1.0`; `pho` resolves the tag
to a commit and records that commit in `photon.lock`. A consumer may also pin a
branch (`@main`) or a raw commit, but tags give reproducible builds.

To release a new version, bump the tag (`v0.2.0`), rebuild the `.lumc`, and push
the new tag. Never move an existing tag: consumers who pinned it would silently
get different code.

### Testing locally before publishing

You do not need to push to test a package. Point `pho` at a local path (which
must be a git repository with at least one commit):

```bash
pho add /path/to/mylib@v0.1.0      # local spec, ref optional
pho build
```

Or bypass `pho` entirely and hand the compiler the search path:

```bash
luminalang -I /path/to/mylib/src -I /path/to/mylib/include app.lum
```

or set `LUMINA_PATH` once (entries separated by `;` on Windows, `:` elsewhere):

```bash
LUMINA_PATH="/path/to/mylib/src:/path/to/other/include" luminalang app.lum
```

A local dependency copies your **working tree** (not the tagged commit), so
uncommitted changes are picked up — which is convenient while developing.

### Housekeeping

A `.gitattributes` keeps byte-stable sources so the shipped `.lumc` stays valid:

```gitattributes
* text=auto eol=lf
*.lumc binary
```

A minimal `.gitignore`:

```gitignore
*.exe
*.o
*.obj
*.ll
```

Also add a `LICENSE` and a `README.md` describing what your package exports and
how to use it (a table of `fn` signatures works well).

### Checklist

- [ ] `lumina.toml` with a unique `name` and `[lib] modules`
- [ ] Source under `src/<module>.lum`, module name unique and ideally equal to
      the package name
- [ ] `.gitattributes` normalising to LF, `.gitignore`, `LICENSE`, `README.md`
- [ ] `luminalang -c src/<module>.lum -o src/<module>.lumc` run, with the
      `.lumi` moved to `include/`
- [ ] Local test from a scratch project with `pho add /path/to/pkg` + `pho build`
- [ ] Committed, tagged (`vX.Y.Z`), and `git push origin main vX.Y.Z`
- [ ] Smoke-test the published version with `pho add github:YOU/pkg@vX.Y.Z`

### Worked example

A complete, minimal package repository:

```
mylib/
├── .gitattributes
├── .gitignore
├── LICENSE                          # Apache-2.0
├── README.md
├── lumina.toml
├── include/mylib.lumi               # module mylib / fn square(i32) i32 / ...
└── src/
    ├── mylib.lum
    └── mylib.lumc
```

`lumina.toml`:

```toml
[package]
name = "mylib"
version = "0.1.0"
description = "A minimal LuminaLang package."
license = "Apache-2.0"
authors = ["Your Name"]

[lib]
modules = ["mylib"]

[dependencies]
```

`src/mylib.lum`:

```lum
fn square(n: i32) i32 {
    back n * n
}

fn answer() i32 {
    back 42
}
```

Publish:

```bash
luminalang -c src/mylib.lum -o src/mylib.lumc
mv src/mylib.lumi include/mylib.lumi
git add -A && git commit -m "new:mylib 0.1.0"
git tag v0.1.0
git remote add origin https://github.com/YOU/mylib.git
git push -u origin main
git push origin v0.1.0
```

Consume:

```bash
pho init app
pho add github:YOU/mylib@v0.1.0
```

```lum
// app/src/main.lum
use mylib::square, answer

fn main() i32 {
    out square(6)     // 36
    out answer()      // 42
    back 0
}
```

[回到顶部 ↑](#)

---

## Chinese

> 简体中文版（以下为同一份开发指南的中文版本）

本指南面向想要**发布 LuminaLang 第三方库**的开发者——确切地说，是发布一个可供其他 LuminaLang 项目通过包管理器 [`pho`](README.md#packages-photon) 安装的库。

photon 是去中心化的：一个包就是一个 git 仓库（通常在 GitHub 上）。没有需要注册的中央仓库，也没有账号要建——你只需要推送一个仓库并打上版本标签即可发布。

### 快速开始

```bash
# 1. 手工创建仓库布局
mkdir mylib && cd mylib
git init -b main
mkdir src include

# 2. 编写清单与源码
#    （见下文各小节）

# 3. 生成预编译库与声明文件
luminalang -c src/mylib.lum -o src/mylib.lumc
mv src/mylib.lumi include/mylib.lumi

# 4. 提交、打标签、推送
git add -A
git commit -m "new:mylib 0.1.0"
git tag v0.1.0
git remote add origin https://github.com/YOU/mylib.git
git push -u origin main
git push origin v0.1.0
```

使用方随后执行：

```bash
pho add github:YOU/mylib@v0.1.0
```

并用 `use mylib::something` 调用你的函数。

### 包的目录结构

一个仓库对应一个包，一个包可以包含多个模块。惯例布局如下：

```
mylib/                     # = git 仓库
├── lumina.toml            # 清单（必需）
├── LICENSE                # 建议
├── README.md              # 建议
├── include/
│   └── mylib.lumi         # 纯声明签名（自动生成）
└── src/
    ├── mylib.lum          # 源码模块（必需）
    └── mylib.lumc         # 预编译库（自动生成，可选）
```

严格必需的文件只有 `lumina.toml` 和源码模块。`include/` 与 `.lumc` 让使用方安装更快、包更自解释；但没有它们，使用方也可以直接编译你的 `.lum` 源码。

### lumina.toml 参考

`lumina.toml` 是手工编写的一小份 TOML 子集：支持段、带引号的字符串、字符串数组以及 `#` 注释。未知的键与段会被忽略，所以文件是向前兼容的。

```toml
[package]
name = "mylib"              # 包名（必需）
version = "0.1.0"           # 仅作信息展示；使用方真正钉住的是 git 标签
description = "One line about mylib."   # 可选
license = "Apache-2.0"      # 可选，SPDX 标识
authors = ["Your Name"]     # 可选

[lib]
modules = ["mylib"]         # 包导出的模块（信息性）

[dependencies]
util = "github:someone/util@v1.2.0"     # 零个或更多依赖
```

说明：

- `name` 是包名。它就是 `pho` 写进使用方清单的键，也是 `packages/` 下的目录名，所以要保证唯一。
- `version` 仅作展示——可复现安装靠的是 git 标签（见[版本与发布](#版本与发布)）。
- `[lib] modules` 记录你导出什么，方便人和工具阅读。编译器**并不依赖它**：它靠搜索路径上的 `mylib.lumc`、`mylib.lum` 或 `mylib.lumi` 来定位模块。
- `[dependencies]` 可省略；见[依赖](#依赖)。

### 编写模块

模块就是一个普通的 `.lum` 文件。对已发布库而言，有几个规则很重要：

- **文件名就是模块名。** `src/mylib.lum` 以 `use mylib::...` 被导入。除非有特殊原因，否则让模块名与包名一致。
- **所有顶层 `fn` 都是公开的。** 语言没有 `private` 关键字，你写出来的任何函数都可被使用方调用。要么把辅助函数命名清楚，要么接受它们是 API 的一部分。
- **顶层只允许 `use` 与 `fn`。** 没有顶层可执行代码；常量放在函数内部。
- **`main` 会被忽略**——当文件按库的方式编译时，所以你可以留一个 `main` 用于本地测试。

```lum
// src/mylib.lum
use math::sqrt          // 你可以使用标准库

fn square(n: i32) i32 {
    back n * n
}

fn distance(x1: i32, y1: i32, x2: i32, y2: i32) i32 {
    back sqrt(square(x2 - x1) + square(y2 - y1))
}
```

跨模块边界可用的是语言本身的类型：`i32`、`f64`、`bool`、`string` 以及切片（`[]i32`、`[]f64`、`[]bool`、`[]string`）。

如果包需要标准库，直接 `use` 即可——使用方的构建会提供它；你的 `.lumc` 只按名字引用标准库符号，因此保持目标无关。

### 模块名与 `use` 名

这是最微妙的一条规则，值得读两遍。

使用方**按模块名**（源码文件名）导入，而不是按他们清单里的依赖名。依赖名只决定安装目录（`packages/<name>/`）和锁文件条目。

```
consumer/lumina.toml:   mylib = "github:YOU/mylib@v0.1.0"
consumer puts it at:    packages/mylib/
consumer imports with:  use mylib::distance      # <- 来自 src/mylib.lum
```

由此得出的结论：

- **让模块名保持唯一。** 两个已安装的包如果都带有 `util.lum` 就会产生歧义：编译器会命中搜索路径上先碰到的那个。
- **别名只改目录名，不改模块名。** 如果使用方写 `pho add github:bob/util --name util_b`，包会装到 `packages/util_b/`，但模块名仍然是 `util`——所以一个项目不应安装两个*模块名*相同的包。请选择有辨识度的名字。

### 依赖

一个包可以依赖其他包。在 `[dependencies]` 里声明，写法与使用方使用的 spec 形式相同：

```toml
[dependencies]
util = "github:someone/util@v1.2.0"
```

使用方安装你的包时，`pho` 会一并解析你的依赖：

- 每个 spec 是 `github:owner/repo[@ref]`、`owner/repo[@ref]`（GitHub）、`https://host/owner/repo[@ref]`、`git@host:owner/repo[@ref]` 或本地路径中的一种。
- `@ref` 是标签、分支或 commit。省略则跟随默认分支。
- 解析是递归的，结果写入使用方的 `photon.lock`，把每个包钉到精确的 commit。
- **重名会大声失败。** 如果依赖图中有两个包对同一依赖名要求了不同来源，安装会直接报错，而不是猜测。

想在自己的代码里 `use` 某个依赖的模块，直接写 `use <模块名>::...` 即可——所有已安装包的 `src/`、`include/` 与根目录都在共享的模块搜索路径上。

### 声明文件（`.lumi`）

`.lumi` 文件是模块公开签名的纯文本列表：

```
// mylib.lumi - generated by LuminaLang; declarations only.
module mylib
fn distance(x1: i32, y1: i32, x2: i32, y2: i32) i32
fn square(n: i32) i32
```

- `luminalang -c` 会在输出旁自动生成它；按惯例把它移到 `include/` 下。
- 它是兜底方案：当找不到 `.lumc` 或 `.lum` 时，编译器只加载 `.lumi`，并对声明的签名做类型检查，实现在链接期提供。
- 它便于编辑器使用，也能在隐藏实现的情况下发布"头文件"。

### 预编译库（`.lumc`）

`.lumc` 文件是预编译模块：签名表加上模块的目标无关 LLVM IR。存在时，使用方直接链接它，而不重新解析你的源码。

```bash
luminalang -c src/mylib.lum -o src/mylib.lumc
```

- **你可以提交 `.lumc`。** 它让安装/构建更快，也能在不暴露 `.lum` 源码的情况下发布包。提交它是安全的，因为它携带了构建它的源码的哈希。
- **过期会被自动处理。** 如果同一个 `.lum` 与 `.lumc` 并排放置但已不匹配记录的哈希，编译器会忽略过期的 `.lumc` 并改用源码编译。所以修改模块永远会生效。
- **发布时要重新生成。** 再跑一次 `-c`，保证提交的 `.lumc` 与打标签的源码一致。
- **换行符很关键。** 哈希是对源码字节计算的，所以要在 `.gitattributes` 里把 `.lum`/`.lumi` 规范为 LF（见[仓库卫生](#仓库卫生)）；否则 CRLF 检出会让 `.lumc` 失效，使用方会静默退回源码编译（结果仍正确，只是更慢）。

### 版本与发布

版本就是普通的 git 标签，没有独立的发布流程。

1. 挑选标签名，如 `v0.1.0`（SemVer 是惯例，但不强制）。
2. 提交最终、已刷新 `.lumc` 的树。
3. 打标签并推送该标签：

   ```bash
   git tag v0.1.0
   git push origin main
   git push origin v0.1.0
   ```

使用方用 `pho add github:YOU/mylib@v0.1.0` 钉住它；`pho` 把标签解析成 commit 并记录进 `photon.lock`。使用方也可以钉分支（`@main`）或裸 commit，但标签才能保证可复现构建。

发布新版本时，把标签升一版（`v0.2.0`）、重新生成 `.lumc` 并推送新标签。**永远不要移动已经存在的标签**：钉住它的使用方会悄悄拿到不同的代码。

### 发布前本地测试

测试包不需要先推送。把 `pho` 指向本地路径（它必须是一个至少有一个 commit 的 git 仓库）：

```bash
pho add /path/to/mylib@v0.1.0      # 本地 spec，ref 可选
pho build
```

或者完全绕过 `pho`，直接把搜索路径交给编译器：

```bash
luminalang -I /path/to/mylib/src -I /path/to/mylib/include app.lum
```

或一次性设置 `LUMINA_PATH`（Windows 上用 `;` 分隔条目，其他平台用 `:`）：

```bash
LUMINA_PATH="/path/to/mylib/src:/path/to/other/include" luminalang app.lum
```

本地依赖复制的是你的**工作树**（而非打包的 commit），因此未提交的改动也会被纳入——开发时很方便。

### 仓库卫生

（Housekeeping）`.gitattributes` 保证源码字节稳定，使提交的 `.lumc` 始终有效：

```gitattributes
* text=auto eol=lf
*.lumc binary
```

一份最小 `.gitignore`：

```gitignore
*.exe
*.o
*.obj
*.ll
```

同时加一个 `LICENSE` 和描述包导出内容与用法的 `README.md`（列一张 `fn` 签名表就很合适）。

### 检查清单

- [ ] `lumina.toml`：唯一的 `name` 和 `[lib] modules`
- [ ] 源码在 `src/<module>.lum`，模块名唯一且最好与包名一致
- [ ] `.gitattributes`（规范 LF）、`.gitignore`、`LICENSE`、`README.md`
- [ ] 运行过 `luminalang -c src/<module>.lum -o src/<module>.lumc`，并把 `.lumi` 移到 `include/`
- [ ] 用 `pho add /path/to/pkg` + `pho build` 在临时项目里本地测过
- [ ] 已提交、打标签（`vX.Y.Z`），并 `git push origin main vX.Y.Z`
- [ ] 用 `pho add github:YOU/pkg@vX.Y.Z` 冒烟测试过已发布版本

### 完整示例

一个完整的最小包仓库：

```
mylib/
├── .gitattributes
├── .gitignore
├── LICENSE                          # Apache-2.0
├── README.md
├── lumina.toml
├── include/mylib.lumi               # module mylib / fn square(i32) i32 / ...
└── src/
    ├── mylib.lum
    └── mylib.lumc
```

`lumina.toml`：

```toml
[package]
name = "mylib"
version = "0.1.0"
description = "A minimal LuminaLang package."
license = "Apache-2.0"
authors = ["Your Name"]

[lib]
modules = ["mylib"]

[dependencies]
```

`src/mylib.lum`：

```lum
fn square(n: i32) i32 {
    back n * n
}

fn answer() i32 {
    back 42
}
```

发布：

```bash
luminalang -c src/mylib.lum -o src/mylib.lumc
mv src/mylib.lumi include/mylib.lumi
git add -A && git commit -m "new:mylib 0.1.0"
git tag v0.1.0
git remote add origin https://github.com/YOU/mylib.git
git push -u origin main
git push origin v0.1.0
```

使用：

```bash
pho init app
pho add github:YOU/mylib@v0.1.0
```

```lum
// app/src/main.lum
use mylib::square, answer

fn main() i32 {
    out square(6)     // 36
    out answer()      // 42
    back 0
}
```

[回到顶部 ↑](#)