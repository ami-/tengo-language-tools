local M = {}

---@param opts? {lsp?: table, enable_lsp?: boolean}
M.setup = function(opts)
  opts = opts or {}

  local function resolve_bin(name)
    local from_path = vim.fn.exepath(name)
    if from_path ~= '' then return from_path end
    local lazy_bin = vim.fn.stdpath('data') .. '/lazy/tengo-language-tools/' .. name
    if vim.uv.fs_stat(lazy_bin) then return lazy_bin end
    return name
  end

  vim.lsp.config('tengols', vim.tbl_deep_extend('force', {
    cmd = { resolve_bin('tengols') },
    filetypes = { 'tengo' },
    root_markers = { '.git' },
  }, opts.lsp or {}))

  if opts.enable_lsp then
    vim.lsp.enable('tengols')
  end
end

return M
