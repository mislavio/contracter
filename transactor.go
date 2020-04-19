package main

import (
	"encoding/base64"
	"errors"
	"log"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/upvestco/upvest-go"
)

func newUpvestTransactor(c *upvest.ClienteleAPI) *bind.TransactOpts {
	conf, err := getConfig()

	w, err := c.Wallet.Get(conf.UpvestWalletID)
	if err != nil {
		log.Fatal(err)
	}

	fromAddress := common.HexToAddress(w.Address)

	return &bind.TransactOpts{
		From: fromAddress,
		Signer: func(signer types.Signer, address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != fromAddress {
				return nil, errors.New("not authorized to sign this account")
			}

			sp := &upvest.SignatureParams{
				Password: conf.UpvestPassword,
				ToSign:   base64.StdEncoding.EncodeToString(signer.Hash(tx).Bytes()),
			}

			upvestSignature, err := c.Wallet.Sign(conf.UpvestWalletID, sp)
			if err != nil {
				return nil, err
			}

			// Convert and concat R, S and V components into []byte{}
			r, err := base64.StdEncoding.DecodeString(upvestSignature.R)
			if err != nil {
				return nil, err
			}

			s, err := base64.StdEncoding.DecodeString(upvestSignature.S)
			if err != nil {
				return nil, err
			}

			recover, err := strconv.Atoi(upvestSignature.Recover)
			if err != nil {
				return nil, err
			}

			v := byte(recover)

			signature := append(r, s...)
			signature = append(signature, v)

			return tx.WithSignature(signer, signature)
		},
	}
}
