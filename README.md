# tengo-language-tools

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Developer tooling for the [Tengo](https://github.com/d5/tengo) scripting language.

## Tools

### tengofmt

Formats Tengo source files.

```
Usage: tengofmt [flags] [file...]

  -l int   max line length for inlining map literals (0 to always expand, default 100)
  -w       write result to source file instead of stdout
```

Reads from stdin if no files are given.

**Examples:**

```bash
# format to stdout
tengofmt file.tengo

# format in-place
tengofmt -w file.tengo

# pipe
cat file.tengo | tengofmt
```

### tengols

Language server (LSP) for Tengo. Works with any LSP-compatible editor.

| Capability | Details |
|---|---|
| **Diagnostics** | Syntax errors reported on save/change |
| **Hover** | Signature and doc comment for local functions, imported module members, and all Tengo stdlib functions (`fmt`, `math`, `os`, `text`, `times`, `rand`, `json`, `hex`, `base64`, `enum`) |
| **Go to definition** | Jumps to the definition of variables, functions, function parameters, for-loop variables, and module members; falls back to the deepest resolvable parent in a selector chain |
| **Find references** | All usages of a symbol in the current file |
| **Completion** | Dot-triggered member completion for local modules and stdlib; bare-identifier completion for locals, imports, and keywords |
| **Rename** | Renames a symbol across the current file; for exported symbols, propagates to all workspace files that import the module |
| **Document symbols** | Outline of top-level functions and variables |
| **Formatting** | Full-document format via `tengofmt` |

## Building

```bash
make        # builds tengofmt and tengols into the repo root
make clean  # removes built binaries
```

## Neovim

Requires Neovim 0.11+ and [lazy.nvim](https://github.com/folke/lazy.nvim).

### lazy.nvim

```lua
{
  'ami-/tengo-language-tools',
  build = 'make',
  lazy = false,
  config = function()
    require('tengo-language-tools').setup({
      -- enable_lsp = true,  -- uncomment when tengols is implemented
      formatter = {
        max_line_len = 100,  -- inline map literals up to this width (0 to always expand)
      },
    })
  end,
}
```

The `build = 'make'` step compiles `tengofmt` and `tengols` into the plugin directory.

`setup()` automatically registers `tengofmt` with [conform.nvim](https://github.com/stevearc/conform.nvim) if it is loaded. Add `tengo` to `formatters_by_ft` to activate it:

```lua
-- In your conform.nvim opts:
formatters_by_ft = {
  tengo = { 'tengofmt' },
},
```

## Related

- [tree-sitter-tengo](https://github.com/ami-/tree-sitter-tengo) — Tree-sitter grammar for Tengo (syntax highlighting in Neovim and other editors)
