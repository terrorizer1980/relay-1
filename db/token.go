package db

import (
	"fmt"

	"github.com/cihub/seelog"
	"github.com/eosforce/relay/token"
	"github.com/eosforce/relay/types"
	"github.com/fanyang1988/eos-go"
	"github.com/pkg/errors"
	pg "gopkg.in/pg.v4"
)

// AccountToken Account's Token in db
type AccountToken struct {
	Name       string `json:"name" sql:",pk"`
	Chain      string `json:"chain" sql:",pk"`
	TokenChain string `json:"token_chain" sql:",pk"`
	Symbol     string `json:"symbol" sql:",pk"`
	Amount     int64  `json:"amount"`
}

func (u *AccountToken) String() string {
	return fmt.Sprintf("%s::%s %s::%s",
		u.Chain, u.Name, u.TokenChain, eos.Asset{
			Amount: u.Amount,
			Symbol: eos.Symbol{
				Symbol: u.Symbol,
			}}.String())
}

// GetAccountTokens get account tokens, not REALTIME!!!
func GetAccountTokens(name, chain string) ([]types.Asset, error) {
	var res []AccountToken
	err := Get().Model(&res).Where("name=? and chain=?", name, chain).Select()

	if err != nil {
		return nil, err
	}

	r := make([]types.Asset, 0, len(res))
	for _, rr := range res {
		typ, err := token.GetSymbol(rr.TokenChain, rr.Symbol)
		if err != nil {
			return []types.Asset{},
				seelog.Warnf("get token  err by %s", err)
		}
		r = append(r, types.NewAsset(
			rr.Chain,
			rr.Amount,
			typ.Symbol,
		))
	}

	return r, nil
}

// GetAccountToken get account tokens, not REALTIME!!!
func GetAccountToken(db *pg.DB, name, chain, tokenChain, symbol string) (types.Asset, error) {
	typ, err := token.GetSymbol(chain, symbol)
	if err != nil {
		return types.Asset{}, err
	}

	var rs []AccountToken
	err = Get().Model(&rs).
		Where("name=?,chain=?,token_chain=?,symbol=?",
			name, chain, tokenChain, symbol).
		Select()

	if err != nil || len(rs) == 0 {
		return types.NewAsset(
			tokenChain,
			0,
			typ.Symbol,
		), err
	}

	r := rs[0]

	return types.NewAsset(
		r.Chain,
		r.Amount,
		typ.Symbol,
	), nil
}

// AddToken add token to account
func AddToken(name, chain string, asset types.Asset) (resErr error) {
	db := Get()
	tx, err := db.Begin()
	if err != nil {
		return errors.WithMessage(err, "add token")
	}

	// TODO By FanYang use a simple err process func
	defer func() {
		if err := recover(); err != nil {
			e, ok := err.(error)
			if ok {
				resErr = e
			} else {
				resErr = fmt.Errorf("err by %v", err)
			}
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				resErr = errors.WithStack(
					seelog.Errorf("rollback err by %s when err %s",
						rollbackErr.Error(), resErr.Error()))
			}
			return
		}
	}()

	newToken := AccountToken{
		Name:       name,
		Chain:      chain,
		TokenChain: string(asset.Chain),
		Symbol:     asset.GetSymbol(),
		Amount:     asset.Amount,
	}

	var rs []AccountToken
	err = tx.Model(&rs).
		Where("name=? and chain=? and token_chain=? and symbol=?",
			name, chain, string(asset.Chain), string(asset.Symbol.Symbol.Symbol)).
		Select()
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.WithStack(
				seelog.Errorf("rollback err by %s when err %s",
					rollbackErr.Error(), err.Error()))
		}
		return errors.WithMessage(err, "add token select err ")
	}

	if len(rs) != 0 {
		// only use first
		newToken.Amount += rs[0].Amount
		_, err = tx.Model(&newToken).Update()
	} else {
		_, err = tx.Model(&newToken).Create()
	}

	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.WithStack(
				seelog.Errorf("rollback err by %s when err %s",
					rollbackErr.Error(), err.Error()))
		}
		return errors.WithMessage(err, "create add token")
	}

	return tx.Commit()
}

// CostToken cost token from account if no token return a error
func CostToken(name, chain string, asset types.Asset) error {
	return nil
}
