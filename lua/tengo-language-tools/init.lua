local M = {}

---@param opts? {lsp?: table, enable_lsp?: boolean, formatter?: {max_line_len?: integer}}
M.setup = function(opts)
  opts = opts or {}

  local function resolve_bin(name)
    local from_path = vim.fn.exepath(name)
    if from_path ~= '' then return from_path end
    -- Plugin root is 3 levels up from lua/tengo-language-tools/init.lua
    local plugin_root = vim.fn.fnamemodify(debug.getinfo(1, 'S').source:sub(2), ':h:h:h')
    local plugin_bin = plugin_root .. '/' .. name
    if vim.uv.fs_stat(plugin_bin) then return plugin_bin end
    local lazy_bin = vim.fn.stdpath('data') .. '/lazy/tengo-language-tools/' .. name
    if vim.uv.fs_stat(lazy_bin) then return lazy_bin end
    return name
  end

  -- LSP
  vim.lsp.config('tengols', vim.tbl_deep_extend('force', {
    cmd = { resolve_bin('tengols') },
    filetypes = { 'tengo' },
    root_markers = { '.git' },
  }, opts.lsp or {}))

  if opts.enable_lsp then
    vim.lsp.enable('tengols')
  end

  -- Formatter: register tengofmt with conform.nvim if available.
  local conform_ok, conform = pcall(require, 'conform')
  if conform_ok then
    local fmt_opts = opts.formatter or {}
    local args = {}
    if fmt_opts.max_line_len ~= nil then
      args = { '-l', tostring(fmt_opts.max_line_len) }
    end
    conform.formatters.tengofmt = {
      command = resolve_bin('tengofmt'),
      args = args,
    }
  end
end

return M
