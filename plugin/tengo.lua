-- Auto-register tengols LSP config without enabling it.
-- To enable the LSP, call:
--   require('tengo-language-tools').setup({ enable_lsp = true })
-- or directly:
--   vim.lsp.enable('tengols')
require('tengo-language-tools').setup({ enable_lsp = false })
