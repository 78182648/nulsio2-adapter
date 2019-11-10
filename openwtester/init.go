package openwtester

import (
	"github.com/blocktree/nulsio2-adapter/nulsio2"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openw"
)

func init() {
	//注册钱包管理工具
	log.Notice("Wallet Manager Load Successfully.")
	openw.RegAssets(nulsio2.Symbol, nulsio2.NewWalletManager())
}
