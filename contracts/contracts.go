package contracts

import (
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mislavio/contracter/accounts"
	"github.com/mislavio/contracter/helpers"
)

// Contract represents a smart contract published to the contracter API.
type Contract struct {
	helpers.BaseModel
	ABI      postgres.Jsonb `json:"abi"`
	Bytecode []byte         `json:"bytecode"`
	Address  string         `json:"address"`
}

// MyContract represents the mapping between an Account
// and a linked Contracter platform Contract
type MyContract struct {
	helpers.BaseModel
	AccountID  string
	Account    accounts.Account
	ContractID string
	Contract   Contract
}
