package openwtester

import (
	"github.com/blocktree/openwallet/v2/log"
	"github.com/blocktree/openwallet/v2/openw"
	"github.com/blocktree/openwallet/v2/openwallet"
	"path/filepath"
	"testing"
)

var (
	testApp        = "NULScoin-adapter"
	configFilePath = filepath.Join("conf")
	dbFilePath     = filepath.Join("data", "db")
	dbFileName     = "blockchain-NULS2.db"
)

func testInitWalletManager() *openw.WalletManager {
	log.SetLogFuncCall(true)
	tc := openw.NewConfig()

	tc.ConfigDir = configFilePath
	tc.EnableBlockScan = false
	tc.SupportAssets = []string{
		"NULS2",
	}
	return openw.NewWalletManager(tc)
	//tm.Init()
}

func TestWalletManager_CreateWallet(t *testing.T) {
	tm := testInitWalletManager()
	w := &openwallet.Wallet{Alias: "HELLO NULS2", IsTrust: true, Password: "12345678"}
	nw, key, err := tm.CreateWallet(testApp, w)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("wallet:", nw)
	log.Info("key:", key)

}

func TestWalletManager_GetWalletInfo(t *testing.T) {

	tm := testInitWalletManager()

	wallet, err := tm.GetWalletInfo(testApp, "VzGeU7t6vj2u1dzVmLJrWWsi8DFRBsFAE7")
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}
	log.Info("wallet:", wallet)
}

func TestWalletManager_GetWalletList(t *testing.T) {

	tm := testInitWalletManager()

	list, err := tm.GetWalletList(testApp, 0, 10000000)
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}
	for i, w := range list {
		log.Info("wallet[", i, "] :", w)
	}
	log.Info("wallet count:", len(list))

	tm.CloseDB(testApp)
}

//HhMp9EJwZpNFhfUuSSXanocxgPGz9eLoSbPbqawcWtWU
//Nse5VJW4vNDyJXkcMmCTHRD9QQC5T1WS
func TestWalletManager_CreateAssetsAccount(t *testing.T) {

	tm := testInitWalletManager()

	walletID := "VzGeU7t6vj2u1dzVmLJrWWsi8DFRBsFAE7"
	account := &openwallet.AssetsAccount{Alias: "mainnetNULS4", WalletID: walletID, Required: 1, Symbol: "NULS2", IsTrust: true}
	account, address, err := tm.CreateAssetsAccount(testApp, walletID, "12345678", account, nil)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("account:", account)
	log.Info("address:", address)

	tm.CloseDB(testApp)
}

func TestWalletManager_GetAssetsAccountList(t *testing.T) {

	tm := testInitWalletManager()

	walletID := "VzGeU7t6vj2u1dzVmLJrWWsi8DFRBsFAE7"
	list, err := tm.GetAssetsAccountList(testApp, walletID, 0, 10000000)
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}
	for i, w := range list {
		log.Infof("account[%d] : %+v", i, w)
	}
	log.Info("account count:", len(list))

	tm.CloseDB(testApp)

}

func TestWalletManager_CreateAddress(t *testing.T) {

	tm := testInitWalletManager()

	walletID := "VzGeU7t6vj2u1dzVmLJrWWsi8DFRBsFAE7"
	//accountID := "7AskbZZjhevnJxWAsNZ5HsyeddptEsqVwPuTu9CcpMch"
	accountID := "6Dy7yK3w73CSLtC6QD3ddvxNsVZ9f8ffkzEuzoDi5pt4"
	address, err := tm.CreateAddress(testApp, walletID, accountID, 5)
	if err != nil {
		log.Error(err)
		return
	}

	for _, addr := range address {
		log.Info(addr.Address)
	}

	tm.CloseDB(testApp)
}

func TestWalletManager_GetAddressList(t *testing.T) {

	tm := testInitWalletManager()

	walletID := "VzGeU7t6vj2u1dzVmLJrWWsi8DFRBsFAE7"
	accountID := "7AskbZZjhevnJxWAsNZ5HsyeddptEsqVwPuTu9CcpMch"
	//accountID := "4h4wnCmpzgy3ZTeoMHs3gjDCuWyXQcxDsk9dcwbNGhmR"
	list, err := tm.GetAddressList(testApp, walletID, accountID, 0, -1, false)
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}
	for i, w := range list {
		log.Info("address[", i, "] :", w.Address)
		//log.Info("address[", i, "] :", w.PublicKey)
	}
	log.Info("address count:", len(list))

	tm.CloseDB(testApp)
}

func TestWalletManager_GetAddressList2(t *testing.T) {

	tm := testInitWalletManager()

	walletID := "VzGeU7t6vj2u1dzVmLJrWWsi8DFRBsFAE7"
	accountID := "6Dy7yK3w73CSLtC6QD3ddvxNsVZ9f8ffkzEuzoDi5pt4"
	//accountID := "4h4wnCmpzgy3ZTeoMHs3gjDCuWyXQcxDsk9dcwbNGhmR"
	list, err := tm.GetAddressList(testApp, walletID, accountID, 0, -1, false)
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}
	for i, w := range list {
		log.Info("address[", i, "] :", w.Address)
		//log.Info("address[", i, "] :", w.PublicKey)
	}
	log.Info("address count:", len(list))

	tm.CloseDB(testApp)
}

func TestBatchCreateAddressByAccount(t *testing.T) {

	tm := testInitWalletManager()

	symbol := "NULS2"
	walletID := "VzGeU7t6vj2u1dzVmLJrWWsi8DFRBsFAE7"
	accountID := "7AskbZZjhevnJxWAsNZ5HsyeddptEsqVwPuTu9CcpMch"

	account, err := tm.GetAssetsAccountInfo(testApp, walletID, accountID)
	if err != nil {
		t.Errorf("error: %v", err)
		return
	}

	assetsMgr, err := openw.GetAssetsAdapter(symbol)
	if err != nil {
		log.Error(symbol, "is not support")
		return
	}

	//decoder := assetsMgr.GetAddressDecode()

	addrArr, err := openwallet.BatchCreateAddressByAccount(account, assetsMgr, 10, 20)
	if err != nil {
		t.Errorf("error: %v", err)
		return
	}
	addrs := make([]string, 0)
	for _, a := range addrArr {
		log.Infof("address[%d]: %s", a.Index, a.Address)
		addrs = append(addrs, a.Address)
	}
	log.Infof("create address")

}

func TestBatchCreateAddressByAccount2(t *testing.T) {

	tm := testInitWalletManager()

	symbol := "NULS2"
	walletID := "VzGeU7t6vj2u1dzVmLJrWWsi8DFRBsFAE7"
	accountID := "7AskbZZjhevnJxWAsNZ5HsyeddptEsqVwPuTu9CcpMch"

	account, err := tm.GetAssetsAccountInfo(testApp, walletID, accountID)
	if err != nil {
		t.Errorf("error: %v", err)
		return
	}

	assetsMgr, err := openw.GetAssetsAdapter(symbol)
	if err != nil {
		log.Error(symbol, "is not support")
		return
	}

	//decoder := assetsMgr.GetAddressDecode()

	addrArr, err := openwallet.BatchCreateAddressByAccount(account, assetsMgr, 10, 20)
	if err != nil {
		t.Errorf("error: %v", err)
		return
	}
	addrs := make([]string, 0)
	for _, a := range addrArr {
		log.Infof("address[%d]: %s", a.Index, a.Address)
		addrs = append(addrs, a.Address)
	}
	log.Infof("create address")

}

//VzGeU7t6vj2u1dzVmLJrWWsi8DFRBsFAE7  目标
//NULSd6Hgj7CK5drU8PYGMQtjMjgR9zMZKRKbL
