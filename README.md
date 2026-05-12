# Build instructions
First, install the tree-sitter library per the [documentation](https://tree-sitter.github.io/tree-sitter/using-parsers/1-getting-started.html)

Then, build the langauge server with `go build`.

Make sure you have `imp.so` somewhere on NeoVim's runtime path, for example:
`/home/caleb/.local/share/nvim/site/parser/imp.so`

Finally, put the provided `init.lua` script somewhere in your NeoVim
configuratiton, and make sure you point the `cmd` path to the `imp_lsp` binary
from the `go build` step. For example, my binary lives at
`~/src/imp_lsp/imp_lsp`, so my lsp config looks like this:

```lua
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
```
