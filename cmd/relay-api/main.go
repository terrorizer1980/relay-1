package main

import (
	"flag"

	"github.com/cihub/seelog"
	"github.com/eosforce/relay/cmd/relay-api/http"
	"github.com/eosforce/relay/token"
	"github.com/eosforce/relay/types"
	"github.com/gin-gonic/gin"

	"github.com/eosforce/relay/cmd/config"
	_ "github.com/eosforce/relay/cmd/relay-api/account"
	_ "github.com/eosforce/relay/cmd/relay-api/chain"
	_ "github.com/eosforce/relay/cmd/relay-api/token"
	"github.com/eosforce/relay/db"
)

// relay http api

var url = flag.String("url", ":8080", "api url")
var logCfg = flag.String("logCfg", "", "log xml cfg file path")
var isDebug = flag.Bool("d", false, "is in debug mode")

func main() {
	defer seelog.Flush()
	flag.Parse()

	// TODO use cfg
	db.InitDB(db.PostgresCfg{
		Address:  "127.0.0.1:5432",
		User:     "pgfy",
		Password: "123456",
		Database: "test3",
	})

	// TODO By FanYang user def symbol
	token.Reg(types.Symbol{
		Precision: 4,
		Chain:     "main",
		Symbol:    "EOS",
	})
	token.Reg(types.Symbol{
		Precision: 4,
		Chain:     "side",
		Symbol:    "EOS",
	})
	token.Reg(types.Symbol{
		Precision: 4,
		Chain:     "side",
		Symbol:    "SYS",
	})
	token.Reg(types.Symbol{
		Precision: 4,
		Chain:     "main",
		Symbol:    "TST",
	})

	config.InitLogger("relay_api", *logCfg)

	http.InitRouter(*isDebug)

	http.Router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	err := http.Router.Run(*url)
	if err != nil {
		seelog.Errorf("run err by %s", err.Error())
	}
}
