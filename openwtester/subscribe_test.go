/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package openwtester

import (
	"github.com/astaxie/beego/config"
	"github.com/blocktree/openwallet/v2/common/file"
	"github.com/blocktree/openwallet/v2/log"
	"github.com/blocktree/openwallet/v2/openw"
	"github.com/blocktree/openwallet/v2/openwallet"
	"path/filepath"
	"testing"
	"time"
)

//var tm = testInitWalletManager()

////////////////////////// 测试单个扫描器 //////////////////////////

type subscriberSingle struct {
	manager *openw.WalletManager
}

//BlockScanNotify 新区块扫描完成通知
func (sub *subscriberSingle) BlockScanNotify(header *openwallet.BlockHeader) error {
	log.Notice("header:", header)
	return nil
}

//BlockTxExtractDataNotify 区块提取结果通知
func (sub *subscriberSingle) BlockExtractDataNotify(sourceKey string, data *openwallet.TxExtractData) error {
	log.Notice("account:", sourceKey)

	for i, input := range data.TxInputs {
		log.Std.Notice("data.TxInputs[%d]: %+v", i, input)
	}

	for i, output := range data.TxOutputs {
		log.Std.Notice("data.TxOutputs[%d]: %+v", i, output)
	}

	log.Std.Notice("data.Transaction: %+v", data.Transaction)

	//tm := testInitWalletManager()
	//walletID := "VzLUoGiZioDZDyisPtKFMD7Sfy485Qih2N"
	//accountID := "HhMp9EJwZpNFhfUuSSXanocxgPGz9eLoSbPbqawcWtWU"

	//testGetAssetsAccountBalance(tm, walletID, accountID)
	//walletID := "VzLUoGiZioDZDyisPtKFMD7Sfy485Qih2N"
	//accountID := "CbhEiN6Pm3ZjJDCkwybanzs192Mo32jhph2RY4ZLMAFN"
	//
	//contract := openwallet.SmartContract{
	//	Address:"NseCpCRzVU3U9RSYyTwSFhdL71wEnpDv",
	//	Decimals:8,
	//	Name:"angel",
	//}
	//balance, err := tm.GetAssetsAccountTokenBalance(testApp,walletID, accountID, contract)
	////balance, err := sub.manager.GetAssetsAccountBalance(testApp, walletID, accountID)
	//if err != nil {
	//	log.Error("GetAssetsAccountBalance failed, unexpected error:", err)
	//	return nil
	//}
	//log.Notice("account balance:", balance.Balance.Balance)

	return nil
}

func (sub *subscriberSingle) BlockExtractSmartContractDataNotify(sourceKey string, data *openwallet.SmartContractReceipt) error {

	log.Notice("sourceKey:", sourceKey)
	log.Std.Notice("data.ContractTransaction: %+v", data)

	for i, event := range data.Events {
		log.Std.Notice("data.Events[%d]: %+v", i, event)
	}

	return nil
}

func TestGetTokenBlance(t *testing.T) {
	tm := testInitWalletManager()

	for i := 0; i < 100; i++ {
		walletID := "VzLUoGiZioDZDyisPtKFMD7Sfy485Qih2N"
		accountID := "CbhEiN6Pm3ZjJDCkwybanzs192Mo32jhph2RY4ZLMAFN"

		contract := openwallet.SmartContract{
			Address:  "NseCpCRzVU3U9RSYyTwSFhdL71wEnpDv",
			Decimals: 8,
			Name:     "angel",
		}

		balance, err := tm.GetAssetsAccountTokenBalance(testApp, walletID, accountID, contract)
		if err != nil {
			log.Error("GetAssetsAccountBalance failed, unexpected error:", err)
			return
		}
		log.Notice("account balance:", balance.Balance.Balance)
	}

}

func TestSubscribeAddress(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
		symbol     = "NULS2"
		addrs      = map[string]string{
			"NULSd6HgUoL5aFx8RCMzUjTSDqLWcV2jrg5UM": "NULSd6HgUoL5aFx8RCMzUjTSDqLWcV2jrg5UM",
			"NULSd6HgVkvve8zBsWvu3Vg8RobNtg17XB9nC": "NULSd6HgVkvve8zBsWvu3Vg8RobNtg17XB9nC",
		}
	)

	tm := testInitWalletManager()

	//GetSourceKeyByAddress 获取地址对应的数据源标识
	scanTargetFunc := func(target openwallet.ScanTarget) (string, bool) {
		key, ok := addrs[target.Address]
		if !ok {
			return "", false
		}
		return key, true
	}

	assetsMgr, err := openw.GetAssetsAdapter(symbol)
	if err != nil {
		log.Error(symbol, "is not support")
		return
	}

	//读取配置
	absFile := filepath.Join(configFilePath, symbol+".ini")

	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return
	}
	assetsMgr.LoadAssetsConfig(c)

	assetsLogger := assetsMgr.GetAssetsLogger()
	if assetsLogger != nil {
		assetsLogger.SetLogFuncCall(true)
	}

	scanner := assetsMgr.GetBlockScanner()

	if scanner.SupportBlockchainDAI() {
		file.MkdirAll(dbFilePath)
		dai, err := openwallet.NewBlockchainLocal(filepath.Join(dbFilePath, dbFileName), false)
		if err != nil {
			log.Error("NewBlockchainLocal err: %v", err)
			return
		}

		scanner.SetBlockchainDAI(dai)
	}

	scanner.SetRescanBlockHeight(1422652)

	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}

	scanner.SetBlockScanTargetFunc(scanTargetFunc)

	sub := subscriberSingle{manager: tm}
	scanner.AddObserver(&sub)

	scanner.Run()

	<-endRunning
}

func TestSubscribeScanBlock(t *testing.T) {

	var (
		symbol = "NULS2"
		addrs  = map[string]string{
			//"Nsdzhw5UzcewtbpweAwXvU7MngCmdNag": "sender",
			"NULSd6HgUkssMi6oSjwEn3puNSijLKnyiRV7H": "NULSd6HgUkssMi6oSjwEn3puNSijLKnyiRV7H",
		}
	)

	//GetSourceKeyByAddress 获取地址对应的数据源标识
	scanAddressFunc := func(address string) (string, bool) {
		key, ok := addrs[address]
		if !ok {
			return "", false
		}
		return key, true
	}

	assetsMgr, err := openw.GetAssetsAdapter(symbol)
	if err != nil {
		log.Error(symbol, "is not support")
		return
	}

	//读取配置
	absFile := filepath.Join(configFilePath, symbol+".ini")

	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return
	}
	assetsMgr.LoadAssetsConfig(c)

	assetsLogger := assetsMgr.GetAssetsLogger()
	if assetsLogger != nil {
		assetsLogger.SetLogFuncCall(true)
	}

	//log.Debug("already got scanner:", assetsMgr)
	scanner := assetsMgr.GetBlockScanner()
	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}

	scanner.SetBlockScanAddressFunc(scanAddressFunc)

	sub := subscriberSingle{}
	scanner.AddObserver(&sub)

	time.Sleep(5 * time.Second)
}

func TestExtractTransactionData(t *testing.T) {

	var (
		symbol = "NULS2"
		addrs  = map[string]string{
			//"Nsdzhw5UzcewtbpweAwXvU7MngCmdNag": "sender",
			"NULSd6HgV9sNKEfG6Qti3H6PYfFNKnpUmT2tF": "NULSd6HgV9sNKEfG6Qti3H6PYfFNKnpUmT2tF",
		}
	)

	//GetSourceKeyByAddress 获取地址对应的数据源标识
	scanAddressFunc := func(address string) (string, bool) {
		key, ok := addrs[address]
		if !ok {
			return "", false
		}
		return key, true
	}

	assetsMgr, err := openw.GetAssetsAdapter(symbol)
	if err != nil {
		log.Error(symbol, "is not support")
		return
	}

	//读取配置
	absFile := filepath.Join(configFilePath, symbol+".ini")

	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return
	}
	assetsMgr.LoadAssetsConfig(c)

	assetsLogger := assetsMgr.GetAssetsLogger()
	if assetsLogger != nil {
		assetsLogger.SetLogFuncCall(true)
	}

	//log.Debug("already got scanner:", assetsMgr)
	scanner := assetsMgr.GetBlockScanner()
	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}

	scanner.SetBlockScanAddressFunc(scanAddressFunc)

	resu, _ := scanner.ExtractTransactionData("45551541f7ab98e7ce665b766ac1bce8e305be3e85f4cbe4b1411da1773f2613", func(target openwallet.ScanTarget) (s string, b bool) {
		return "NULSd6HgV9sNKEfG6Qti3H6PYfFNKnpUmT2tF", true
	})

	log.Warn(resu)
}
