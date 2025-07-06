-- This is a sample configuration for integrating youtrack-cli with Neovim/LazyVim.
-- You can add this to your LazyVim configuration files.

-- 1. Keymaps (e.g., in ~/.config/nvim/lua/config/keymaps.lua)
-- If you move the executable to your $PATH, you can change the command to just "youtrack-cli"

local map = vim.keymap.set
local youtrack_executable = "/Users/dukechiu/github/youtrack-cli/youtrack-cli"

-- List YouTrack issues in a terminal
map("n", "<leader>yl", function()
  vim.cmd("split term://" .. youtrack_executable .. " list")
end, { desc = "List YouTrack issues" })

-- Add a work item (opens a terminal to complete the command)
map("n", "<leader>ya", function()
  vim.cmd("split term://" .. youtrack_executable .. " add-work ")
end, { desc = "Add YouTrack work item" })


-- 2. Telescope Integration (e.g., in a new file like ~/.config/nvim/lua/plugins/telescope-youtrack.lua)
-- This requires the telescope.nvim plugin.

local actions = require("telescope.actions")
local finders = require("telescope.finders")
local pickers = require("telescope.pickers")
local conf = require("telescope.config").values

local function list_youtrack_issues_telescope()
  local command = youtrack_executable .. " list --json"
  local handle = io.popen(command)
  if not handle then
    vim.notify("Failed to run youtrack-cli", vim.log.levels.ERROR)
    return
  end

  local json_output = handle:read("*a")
  handle:close()

  local issues = vim.fn.json_decode(json_output)

  if not issues or type(issues) ~= "table" then
    vim.notify("Failed to parse YouTrack issues or no issues found", vim.log.levels.WARN)
    return
  end

  pickers.new({}, {
    prompt_title = "YouTrack Issues",
    finder = finders.new_table {
      results = issues,
      entry_maker = function(entry)
        return {
          value = entry.idReadable,
          display = entry.idReadable .. ": " .. entry.summary,
          ordinal = entry.idReadable .. " " .. entry.summary,
        }
      end,
    },
    sorter = conf.generic_sorter({}),
    attach_mappings = function(prompt_bufnr, map)
      actions.select_default:replace(function()
        actions.close(prompt_bufnr)
        local selection = require("telescope.actions.state").get_selected_entry()
        if selection then
          vim.cmd("split term://" .. youtrack_executable .. " add-work " .. selection.value .. " ")
        end
      end)
      return true
    end,
  }):find()
end

-- Command to trigger the Telescope picker
vim.api.nvim_create_user_command("TelescopeYoutrack", list_youtrack_issues_telescope, {})

-- Keymap for the Telescope picker
map("n", "<leader>yL", "<cmd>TelescopeYoutrack<cr>", { desc = "List YouTrack issues in Telescope" })
