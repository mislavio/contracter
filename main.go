package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mislavio/contracter/accounts"
	"github.com/mislavio/contracter/auth"
	"github.com/mislavio/contracter/contracts"
	"github.com/rs/cors"
	"gopkg.in/yaml.v2"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

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

const listenPort int = 8000

var jwtauth *auth.ContracterJWT

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

func deployContract() (string, string) {
	conf, err := getConfig()
	if err != nil {
		log.Panic(err)
	}

	ethClient, err := ethclient.Dial("https://ropsten.infura.io/v3/" + conf.InfuraProjectID)
	if err != nil {
		log.Panic(err)
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
		log.Panic(err)
	}

	fromAddress := common.HexToAddress(w.Address)

	nonce, err := ethClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Panic(err)
	}

	gasPrice, err := ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		log.Panic(err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(conf.SmartContractABI))
	if err != nil {
		log.Panic(err)
	}

	auth := newUpvestTransactor(clienteleClient)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	address, tx, _, err := bind.DeployContract(auth, parsedABI, common.FromHex(conf.SmartContractBytecode), ethClient, "1.0")
	if err != nil {
		log.Panic(err)
	}

	return address.Hex(), tx.Hash().Hex()

}

func writeJSONResponse(w http.ResponseWriter, content []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(content)
}

func init() {
	jwtauth = &auth.ContracterJWT{SigningKey: []byte("very_secret_secret"), Signer: jwt.SigningMethodHS256}
}

func main() {
	db, err := gorm.Open("postgres", "host=localhost port=5432 user=mislav dbname=contracter sslmode=disable")
	if err != nil {
		log.Print(err)
	}
	defer db.Close()
	db.LogMode(true)

	if err := db.DB().Ping(); err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(
		&accounts.Account{},
		&auth.Credentials{},
		&auth.OAuthCredentials{},
		&contracts.Contract{},
		&contracts.MyContract{},
	)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Mount("/auth", auth.Router(db, jwtauth))

	r.Group(func(r chi.Router) {
		r.Use(auth.Verifier(jwtauth))
		r.Use(auth.AccountAuthenticator)

		r.Post("contracts/deploy", func(w http.ResponseWriter, r *http.Request) {
			address, hash := deployContract()
			body := fmt.Sprintf("The address of the contract is: \n%v\n\nThe transaction hash is: \n%v\n", address, hash)
			w.Write([]byte(body))
		})
	})

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	log.Printf("listening on %d", listenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", listenPort), c.Handler(r)))
}
