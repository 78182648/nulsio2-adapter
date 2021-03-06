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
	"github.com/blocktree/openwallet/v2/openw"
	"testing"
	"time"

	"github.com/blocktree/openwallet/v2/log"
	"github.com/blocktree/openwallet/v2/openwallet"
)

func testGetAssetsAccountBalance(tm *openw.WalletManager, walletID, accountID string) {
	balance, err := tm.GetAssetsAccountBalance(testApp, walletID, accountID)
	if err != nil {
		log.Error("GetAssetsAccountBalance failed, unexpected error:", err)
		return
	}
	log.Error("balance:", balance)
}

func testGetAssetsAccountTokenBalance(tm *openw.WalletManager, walletID, accountID string, contract openwallet.SmartContract) {
	balance, err := tm.GetAssetsAccountTokenBalance(testApp, walletID, accountID, contract)
	if err != nil {
		log.Error("GetAssetsAccountTokenBalance failed, unexpected error:", err)
		return
	}
	log.Info("token balance:", balance.Balance)
}

func testCreateTransactionStep(tm *openw.WalletManager, walletID, accountID, to, amount, feeRate string, contract *openwallet.SmartContract) (*openwallet.RawTransaction, error) {

	//err := tm.RefreshAssetsAccountBalance(testApp, accountID)
	//if err != nil {
	//	log.Error("RefreshAssetsAccountBalance failed, unexpected error:", err)
	//	return nil, err
	//}

	rawTx, err := tm.CreateTransaction(testApp, walletID, accountID, amount, to, feeRate, "", contract)

	if err != nil {
		log.Error("CreateTransaction failed, unexpected error:", err)
		return nil, err
	}

	return rawTx, nil
}

func testCreateSummaryTransactionStep(
	tm *openw.WalletManager,
	walletID, accountID, summaryAddress, minTransfer, retainedBalance, feeRate string,
	start, limit int,
	contract *openwallet.SmartContract,
	feeSupportAccount *openwallet.FeesSupportAccount) ([]*openwallet.RawTransactionWithError, error) {

	rawTxArray, err := tm.CreateSummaryRawTransactionWithError(testApp, walletID, accountID, summaryAddress, minTransfer,
		retainedBalance, feeRate, start, limit, contract, feeSupportAccount)

	if err != nil {
		log.Error("CreateSummaryTransaction failed, unexpected error:", err)
		return nil, err
	}

	return rawTxArray, nil
}

func testSignTransactionStep(tm *openw.WalletManager, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	_, err := tm.SignTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, "12345678", rawTx)
	if err != nil {
		log.Error("SignTransaction failed, unexpected error:", err)
		return nil, err
	}

	log.Infof("rawTx: %+v", rawTx)
	return rawTx, nil
}

func testVerifyTransactionStep(tm *openw.WalletManager, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	//log.Info("rawTx.Signatures:", rawTx.Signatures)

	_, err := tm.VerifyTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, rawTx)
	if err != nil {
		log.Error("VerifyTransaction failed, unexpected error:", err)
		return nil, err
	}

	log.Infof("rawTx: %+v", rawTx)
	return rawTx, nil
}

func testSubmitTransactionStep(tm *openw.WalletManager, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	tx, err := tm.SubmitTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, rawTx)
	if err != nil {
		log.Error("SubmitTransaction failed, unexpected error:", err)
		return nil, err
	}

	log.Std.Info("tx: %+v", tx)
	log.Info("wxID:", tx.WxID)
	log.Info("txID:", rawTx.TxID)

	return rawTx, nil
}

func TestTransfer(t *testing.T) {

	tm := testInitWalletManager()
	walletID := "VzGeU7t6vj2u1dzVmLJrWWsi8DFRBsFAE7"
	accountID := "7AskbZZjhevnJxWAsNZ5HsyeddptEsqVwPuTu9CcpMch"
	//to := "NULSd6HgVkvve8zBsWvu3Vg8RobNtg17XB9nC"

	tos := []string{
		"NULSd6HgjWMfZwMW27op6BP1567Uu4qe6KugD",
		//"NULSd6HgUoL5aFx8RCMzUjTSDqLWcV2jrg5UM",
		//"NULSd6HgYnSuKoZxZzL8ZxvbB6oS2zzqEP2W1",
		//"NULSd6HgYuJYUTZJXscSutnXBp37YUKY3khDf",
		//"NULSd6HgaZoBqc2uxrYF8ApM6f5ZVbcPshzTJ",
		//"NULSd6Hgap7WJw6inBDSccToUdpth8uXzPfvL",
		//"NULSd6HgapiuG6RxP7LXpNa94cCwuaesMQUk8",
	}

	//accountID := "4h4wnCmpzgy3ZTeoMHs3gjDCuWyXQcxDsk9dcwbNGhmR"
	//to := "fiiimYt7qZekpQKZauBGxv8kGFJGdMyYtzSgdP"

	testGetAssetsAccountBalance(tm, walletID, accountID)

	for _, to := range tos {
		rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "0.001", "", nil)
		if err != nil {
			return
		}

		_, err = testSignTransactionStep(tm, rawTx)
		if err != nil {
			return
		}

		_, err = testVerifyTransactionStep(tm, rawTx)
		if err != nil {
			return
		}

		_, err = testSubmitTransactionStep(tm, rawTx)
		if err != nil {
			return
		}
		//time.Sleep(30 * time.Second)
	}

}

func TestTransferNrc20(t *testing.T) {

	address := []string{
		"NULSd6HgYWfCYbxeVLC3zTfcqhtovXZfLY7z1",
		//"NULSd6HgUoL5aFx8RCMzUjTSDqLWcV2jrg5UM",
		//"NULSd6HgYnSuKoZxZzL8ZxvbB6oS2zzqEP2W1",
		//"NULSd6HgYuJYUTZJXscSutnXBp37YUKY3khDf",
		//"NULSd6HgaZoBqc2uxrYF8ApM6f5ZVbcPshzTJ",
		//"NULSd6Hgap7WJw6inBDSccToUdpth8uXzPfvL",
		//"NULSd6HgapiuG6RxP7LXpNa94cCwuaesMQUk8",
	}

	tm := testInitWalletManager()
	walletID := "VzGeU7t6vj2u1dzVmLJrWWsi8DFRBsFAE7"
	accountID := "7AskbZZjhevnJxWAsNZ5HsyeddptEsqVwPuTu9CcpMch"

	contract := &openwallet.SmartContract{
		Address:  "NULSd6HgmBys1gA2SztAKMVfob3NbC2a9iY7T",
		Decimals: 8,
		Name:     "CC",
	}

	for _, a := range address {
		rawTx, err := testCreateTransactionStep(tm, walletID, accountID, a, "1", "", contract)
		if err != nil {
			log.Error(err)
			return
		}

		_, err = testSignTransactionStep(tm, rawTx)
		if err != nil {
			return
		}

		_, err = testVerifyTransactionStep(tm, rawTx)
		if err != nil {
			return
		}

		_, err = testSubmitTransactionStep(tm, rawTx)
		if err != nil {
			return
		}
	}

}

func TestSummary(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "VzGeU7t6vj2u1dzVmLJrWWsi8DFRBsFAE7"
	//accountID := "HhMp9EJwZpNFhfUuSSXanocxgPGz9eLoSbPbqawcWtWU"
	accountID := "6Dy7yK3w73CSLtC6QD3ddvxNsVZ9f8ffkzEuzoDi5pt4"
	summaryAddress := "NULSd6HgVkvve8zBsWvu3Vg8RobNtg17XB9nC"

	testGetAssetsAccountBalance(tm, walletID, accountID)

	rawTxArray, err := testCreateSummaryTransactionStep(tm, walletID, accountID,
		summaryAddress, "", "", "",
		0, 100, nil, nil)
	if err != nil {
		log.Errorf("CreateSummaryTransaction failed, unexpected error: %v", err)
		return
	}

	//执行汇总交易
	for _, rawTxWithErr := range rawTxArray {

		if rawTxWithErr.Error != nil {
			log.Error(rawTxWithErr.Error.Error())
			continue
		}

		_, err = testSignTransactionStep(tm, rawTxWithErr.RawTx)
		if err != nil {
			return
		}

		_, err = testVerifyTransactionStep(tm, rawTxWithErr.RawTx)
		if err != nil {
			return
		}

		_, err = testSubmitTransactionStep(tm, rawTxWithErr.RawTx)
		if err != nil {
			return
		}
	}

}

func TestTokenSummary(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "VzGeU7t6vj2u1dzVmLJrWWsi8DFRBsFAE7"
	accountID := "6Dy7yK3w73CSLtC6QD3ddvxNsVZ9f8ffkzEuzoDi5pt4"
	summaryAddress := "NULSd6Hgj7CK5drU8PYGMQtjMjgR9zMZKRKbL"

	testGetAssetsAccountBalance(tm, walletID, accountID)

	for {
		contract := &openwallet.SmartContract{
			Address:  "NULSd6HgmBys1gA2SztAKMVfob3NbC2a9iY7T",
			Decimals: 8,
			Name:     "CC",
		}

		fee := &openwallet.FeesSupportAccount{
			AccountID:        "7AskbZZjhevnJxWAsNZ5HsyeddptEsqVwPuTu9CcpMch",
			FixSupportAmount: "0.2",
		}

		//feeSupport, _ := tm.GetAssetsAccountInfo(testApp,walletID,"HhMp9EJwZpNFhfUuSSXanocxgPGz9eLoSbPbqawcWtWU")

		rawTxArray, err := testCreateSummaryTransactionStep(tm, walletID, accountID,
			summaryAddress, "", "", "",
			0, 100, contract, fee)
		if err != nil {
			log.Errorf("CreateSummaryTransaction failed, unexpected error: %v", err)
			return
		}

		//执行汇总交易
		for _, rawTxWithErr := range rawTxArray {

			if rawTxWithErr.Error != nil {
				log.Error(rawTxWithErr.Error.Error())
				continue
			}

			_, err = testSignTransactionStep(tm, rawTxWithErr.RawTx)
			if err != nil {
				return
			}

			_, err = testVerifyTransactionStep(tm, rawTxWithErr.RawTx)
			if err != nil {
				return
			}

			_, err = testSubmitTransactionStep(tm, rawTxWithErr.RawTx)
			if err != nil {
				return
			}
		}
		time.Sleep(10 * time.Second)
	}

}

func TestR(t *testing.T) {
	s := "123456789"
	log.Warn(s[len(s)-5:])
}
