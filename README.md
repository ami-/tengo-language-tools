# tengo-language-tools

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

### tengols _(planned)_

Language server (LSP) for Tengo. Will provide diagnostics, hover, go-to-definition, and formatting via any LSP-compatible editor.

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
