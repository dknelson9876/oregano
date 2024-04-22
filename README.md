# Oregano

> My attempt at using Plaid to build a budgeting app

## Origin

Since Intuit's Mint is shutting down, I haven't found a replacement that meets my requirements of:

1. Free
2. Self Hostable
3. Automatic Transaction Sync
4. Rollover Budgets

So this project is my "Fine, I'll do it myself" response that might die before it gets anywhere, guess we'll have to see. And because I like to torture myself with new things, I figured I'd learn how to write TUI programs and learn Go at the same time.

## Usage

The folder that oregano uses defaults to is `~/.config/oregano`. This can be overriden by setting the environment variable `OREGANO_DIR`. Additionally, the current folder will be checked for `config.json` before the configured directory.

Available commands:
```
oregano-cli - Terminal budgeting app
Commands:
* help (h)              Print this menu
* quit (q)              Quit oregano
* link                  Link a new institution (Opens in a new browser tab)
* list (ls)             List accounts or transactions
* alias [id] [alias]    Assign [alias] as the new alias for [id]
* remove (rm) [alias/id...]     Remove a linked institution
* account (acc) [alias/id...]   Print details about specific account(s)
* transactions (trs) [alias/id]  List transactions from a specific account
* import [filename]      Import transactions from a csv file
* print (p) [argument index]    Print more details about something that was output
* edit (e) [wid]        Edit the fields of a transaction
* repair                Using higher level data as authoritative, correct inconsistencies
* new ...               manually create account or transaction
```

## Attribution

Behavior for using Plaid Link borrowed from [landakram's plaid-cli](https://github.com/landakram/plaid-cli)