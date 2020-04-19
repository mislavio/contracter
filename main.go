package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"gopkg.in/yaml.v2"

	"github.com/upvestco/upvest-go"
)

type configuration struct {
	UpvestPassword        string `yaml:"upvestPassword"`
	UpvestUsername        string `yaml:"upvestUsername"`
	UpvestWalletID        string `yaml:"upvestWalletID"`
	WalletAddress         string `yaml:"walletAddress"`
	UpvestOAuthID         string `yaml:"upvestOAuthID"`
	UpvestOAuthSecret     string `yaml:"upvestOAuthSecret"`
	UpvestBaseURL         string `yaml:"upvestBaseURL"`
	UpvestEtherAssetID    string `yaml:"upvestEtherAssetID"`
	SmartContractABI      string `yaml:"smartContractABI"`
	SmartContractBytecode string `yaml:"smartContractBytecode"`
	InfuraProjectID       string `yaml:"infuraProjectID"`
}

func getConfig() (*configuration, error) {
	var conf configuration

	reader, err := os.Open("config.yaml")
	if err != nil {
		return &conf, err
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return &conf, err
	}

	if err := yaml.Unmarshal(data, &conf); err != nil {
		return &conf, err
	}

	return &conf, nil
}

func main() {
	conf, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	ethClient, err := ethclient.Dial("https://ropsten.infura.io/v3/" + conf.InfuraProjectID)
	if err != nil {
		log.Fatal(err)
	}

	c := upvest.NewClient(conf.UpvestBaseURL, nil)
	c.SetUA("upvest-go/1.0.0")

	clienteleClient := c.NewClientele(
		conf.UpvestOAuthID,
		conf.UpvestOAuthSecret,
		conf.UpvestUsername,
		conf.UpvestPassword,
	)

	w, err := clienteleClient.Wallet.Get(conf.UpvestWalletID)
	if err != nil {
		log.Fatal(err)
	}

	fromAddress := common.HexToAddress(w.Address)

	nonce, err := ethClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(fromAddress.Hex(), nonce)

	gasPrice, err := ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(conf.SmartContractABI))
	if err != nil {
		log.Fatal(err)
	}

	auth := newUpvestTransactor(clienteleClient)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	address, tx, _, err := bind.DeployContract(auth, parsedABI, common.FromHex(conf.SmartContractBytecode), ethClient, "1.0")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("The Address of the conract is: %v\nThe transaction hash is %v\n", address.Hex(), tx.Hash().Hex())

}
