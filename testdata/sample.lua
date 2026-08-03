-- Options
vim.opt.expandtab = true
vim.opt.shiftwidth = 2
vim.opt.listchars:append({ tab = "»·", trail = "·", nbsp = "·" })
vim.g.mapleader = " " -- a trailing comment

--[[ A long comment, which spans
     lines and closes here ]]

local lazypath = vim.fn.stdpath("data") .. "/lazy/lazy.nvim"
require("lazy").setup({
	"tpope/vim-sensible",
	{ "neovim/nvim-lspconfig" },
})
require"telescope".setup{}

local function filetype_autocmd(ft, callback)
	vim.api.nvim_create_autocmd("FileType", {
		pattern = ft,
		callback = callback,
	})
end

local function run_file(key, cmd_template, split_cmd)
	local cmd = cmd_template:gsub("%%", vim.fn.expand("%:p"))
	buf_map(0, "n", key, function()
		vim.cmd(split_cmd)
		vim.cmd("terminal " .. cmd)
	end)
end

filetype_autocmd("sql", function()
	run_file("<Leader>r", "psql -d $(cat .db) -f % | less", "split")
end)

local query = [[SELECT id, name
FROM jobs
WHERE state = 'queued']]

local function stats(t)
	local n, total = 0, 0
	for _, v in ipairs(t) do
		if type(v) == "number" and v ~= nil then
			n = n + 1
			total = total + v
		elseif v == false or v == nil then
			goto continue
		end
		::continue::
	end
	while n > 0 do
		n = n - 1
	end
	repeat
		total = total / 2
	until total < 1
	return { count = n, mean = total / math.max(n, 1) }
end

local numbers = { 0xff, 1.5e3, .5, 100, -7 }
local escaped = "a \"quoted\" word and a backslash \\"
local single = 'single quotes too'
print(tostring(#numbers), string.format("%d", 1), escaped, single, query, stats)
