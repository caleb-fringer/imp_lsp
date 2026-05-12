-- Add filetype extension
vim.filetype.add({
    extension = { imp = 'imp' },
})
-- Add imp_lsp config
vim.lsp.config['imp_lsp'] = {
    cmd = {
        -- Path to imp_lsp binary, as created by `go build`
        vim.fs.abspath('~/src/imp_lsp/imp_lsp'),
    },
    -- Must match the filetype extension added above.
    filetypes = { 'imp' },
    root_markers = { '.git' },
}
-- Enable the lsp.
-- Check your version of NeoVim, as the API has changed recently.
-- This was tested with v0.11 and v0.12
vim.lsp.enable('imp_lsp')
