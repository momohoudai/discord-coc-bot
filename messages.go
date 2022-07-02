package main

import (
  "strings"
)

const MSG_VERSION = "I'm CocBot v7.0.0, written in Golang!"
const MSG_RESIST_THINK = "Let's see...\nActive: **%d**\nPassive: **%d**\n"
const MSG_RESIST_AUTO_PASS = "The result is an **Automatic Success**!! \\(^o^)/"
const MSG_RESIST_AUTO_FAIL = "The result is an **Automatic Failure**!! (´・ω・`)"
const MSG_RESIST_NORMAL = "The result is **%d ** !! （｀・ω・´）"
const MSG_ADD_ALIAS_TARGET_NOT_FOUND = "Target not found! Are you sure the target name is correct?"
const MSG_ADD_ALIAS_PASS = "Alias added! **%s** is now also known as **%s**!"
const MSG_ADD_ALIAS_DUPLICATE_FOUND = "Duplicate alias **%s** found! Please remove first with the *alias remove* command"
const MSG_GET_ALIAS_PASS = "**%s** is also known as **%s**!"
const MSG_GET_ALIAS_FAIL = "Sorry, I can't find an alias for **%s**..."
const MSG_REMOVE_ALIAS_PASS = "Done! **%s** is not longer an alias! ^^b"
const MSG_REMOVE_ALIAS_FAIL = "Sorry, I can't find an alias named **%s**..."
const MSG_FIND_PASS_WITH_ALIAS = "I found **%s** (aka **%s**)! ```%s```"
const MSG_FIND_PASS = "I found **%s**! ```%s```"
const MSG_FIND_FAIL = "Sorry...I can't find what you are looking for >_<"
const MSG_HELP_QUERY = "Did you do it correctly? ```%s```"
const MSG_HELP_ADD_ALIAS = "add-alias: Adds an alias to a 'find'\n\t> Usage: !coc add-alias <alias_name> = <target_name>\n\t(if success, you can then do '!coc, find <alias_name>')" 
const MSG_HELP_REMOVE_ALIAS = "remove-alias: Removes an alias\n\t> Usage: !coc remove-alias <alias_name>"
const MSG_HELP_GET_ALIAS = "get-alias: Displays an alias\n\t> Usage: !coc get-alias <alias_name>"
const MSG_RESIST_HELP = "resist: Check CoC resistance!\n\t> Usage: !coc resist <active> vs <passive>"
const MSG_HELP_FIND =  "find: Use this command to find something\n\t> Usage: !coc find <something>" 
var MSG_HELP = strings.Join([]string{"```", MSG_RESIST_HELP, MSG_HELP_FIND, MSG_HELP_GET_ALIAS, MSG_HELP_ADD_ALIAS, MSG_HELP_REMOVE_ALIAS, "```"}, "\n")
const MSG_GENERIC_FAIL = "Sorry, something went wrong...contact Momo? (´・ω・`)"
