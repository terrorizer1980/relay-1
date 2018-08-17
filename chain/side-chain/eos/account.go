package eos

import (
	"github.com/cihub/seelog"
	"github.com/eosforce/relay/db"
	"github.com/fanyang1988/eos-go"
)

// Msg Handler About Account

// onMapAccount map account from side chain
func (h *Handler) onMapAccount(account eos.AccountName) error {
	seelog.Debugf("OnMapAccount %s %s", h.Name(), string(account))
	return db.CreateAccount(string(account), h.Name())
}