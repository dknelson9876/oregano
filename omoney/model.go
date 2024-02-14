package omoney

import (
	"fmt"
)

type Model struct {
	FilePath  string
	Accounts map[string]Account
	Aliases  map[string]string // alias -> uuid
}

func (m *Model) GetAccount(input string) (Account, error) {
	if acc, ok := m.Accounts[input]; ok {
		return acc, nil
	} else if id, ok := m.Aliases[input]; ok {
		return m.Accounts[id], nil
	} else {
		return Account{}, fmt.Errorf("input not recognized as valid account ID or alias: %s", input)
	}
}

func (m *Model) IsValidAccount(input string) bool {
	if _, ok := m.Accounts[input]; ok {
		return true
	} else if _, ok := m.Aliases[input]; ok {
		return true
	} else {
		return false
	}
}

// Given an id, returns that id.
// Given an alias, returns the matching alias.
// So that, given an alias or an id as input, reliably
// change it to an id
func (m *Model) resolveToId(input string) (string, error) {
	if _, ok := m.Accounts[input]; ok {
		return input, nil
	} else if id, ok := m.Aliases[input]; ok {
		return id, nil
	} else {
		return "", fmt.Errorf("input not recognized as valid id or alias")
	}
}

// given a string that is an id or an alias, return the matching
// Account's PlaidToken
func (m *Model) GetAccessToken(input string) (string, error) {
	acc, err := m.GetAccount(input)
	if err != nil {
		return "", err
	} else {
		return acc.PlaidToken, nil
	}
}

func (m *Model) AddAccount(acc Account)  {
	m.Accounts[acc.Id] = acc
	if acc.Alias != "" {
		m.Aliases[acc.Alias] = acc.Id
	}
}

func (m *Model) RemoveAcount(input string) error {
	if acc, ok := m.Accounts[input]; ok {
		delete(m.Accounts, input)
		delete(m.Aliases, acc.Alias)
		return nil
	} else if id, ok := m.Aliases[input]; ok {
		delete(m.Accounts, id)
		delete(m.Aliases, input)
		return nil
	}
	return fmt.Errorf("input not recognized as valid item ID or alias: %s", input)
}

//TODO: be sure to save the model after calling this method
func (m *Model) SetAlias(id string, alias string) error {
	var acc Account
	var ok bool
	if acc, ok = m.Accounts[id]; !ok {
		return fmt.Errorf("account ID `%s` not recognized", id)
	}
	//NOTE: because of how go works, acc is a copy, so we must
	//   assign it back after modifying it
	acc.Alias = alias
	m.Accounts[id] = acc
	m.Aliases[alias] = id
	return nil
}

func (m *Model) AddTransaction(tr Transaction) {
	id := m.Aliases[tr.Account]
	acc := m.Accounts[id]
	acc.AddTransaction(tr)
	m.Accounts[id] = acc
}