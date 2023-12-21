# Oregano

> My attempt at using Plaid to build a budgeting app

Since Intuit's Mint is shutting down, I haven't found a replacement that meets my requirements of:

1. Free
2. Self Hostable
3. Automatic Transaction Sync
4. Rollover Budgets

So this project is my "Fine, I'll do it myself" response that might die before it gets anywhere, guess we'll have to see. And because I like to torture myself with new things, I figured I'd learn how to write TUI programs and learn Go at the same time.

### Home Screen Idea
```
┌───────────┬┬─────────────┬──────────────────────────────────────┐
│           ││   Accounts  │  Budget                              │
│►Home      ││             │                                      │
│ Transact  ││ Saving X    │  ┼ ─────                   ┼         │
│ Net       ││     $1389   │   Fast Food             $30 of $60   │
│ Budget    ││             │  ┼ ──                      ┼         │
│ Bills     ││ Checking Y  │   Shopping              $5 of $20    │
│ Accounts  ││    $35.21   │  ┼ ──────────────          ┼         │
│ Invest    ││             │   Income                $2600 of $52.│
│ Charts?   ││ Saving Y    │  ┼ ─────────               ┼         │
│           ││    $5690    │   Gas                   $47 of $120  │
│           ││             │                                      │
│           ││             │                                      │
│           ││             │                                      │
│           ││             │                                      │
│           ││             │                                      │
│           ││             ├──────────────────────────────────────┤
│           ││             │  Recent Transactions                 │
│           ││             │                                      │
│           ││             │  SPOTIFY XX    S     $15      CHASE  │
│ Settings  ││             │  CHEVRON TAYLS G     $47      DISC   │
│ Oregano   ││             │  PAYCHECK      I     $2600    SAVE   │
└───────────┴┴─────────────┴──────────────────────────────────────┘
```