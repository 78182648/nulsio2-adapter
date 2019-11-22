/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 *
 * The OpenWallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenWallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package nulsio2

import (
	"encoding/hex"
	"fmt"
	"github.com/blocktree/nulsio2-adapter/nulsio2_addrdec"
	"github.com/blocktree/openwallet/common"
	"github.com/blocktree/openwallet/crypto"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/shopspring/decimal"
	"time"
)

// Block model
type Block struct {
	/*
		{
		    "timestamp": "2019-01-24T19:32:05.500",
		    "producer": "blkproducer1",
		    "confirmed": 0,
		    "previous": "0137c066283ef586d4e1dba4711b2ddf0248628595855361d9b0920e7f64ea92",
		    "transaction_mroot": "0000000000000000000000000000000000000000000000000000000000000000",
		    "action_mroot": "60c9f06aef01b1b4b2088785c9239c960bca8fc23cedd6b8104c69c0335a6d39",
		    "schedule_version": 2,
		    "new_producers": null,
		    "header_extensions": [],
		    "producer_signature": "SIG_K1_K11ScNfXdat71utYJtkd8E6dFtvA7qQ3ww9K74xEpFvVCyeZhXTarwvGa7QqQTRw3CLFbsXCsWJFNCHFHLKWrnBNZ66c2m",
		    "transactions": [],
		    "block_extensions": [],
		    "id": "0137c067c65e9db8f8ee467c856fb6d1779dfeb0332a971754156d075c9a37ca",
		    "block_num": 20430951,
		    "ref_block_prefix": 2085023480
		}
	*/
	openwallet.BlockHeader
	Height uint32 `storm:"id"`
	Fork   bool
}


func (n *Block) ToBlockHeader() *openwallet.BlockHeader {
	obj := &openwallet.BlockHeader{}
	//解析json
	obj.Hash = n.Hash
	obj.Merkleroot = n.Merkleroot
	obj.Previousblockhash = n.Previousblockhash
	obj.Height = uint64(n.Height)
	obj.Time = uint64(time.Now().Unix())
	obj.Symbol = "NULS2"
	return obj
}

//UnscanRecord 扫描失败的区块及交易
type UnscanRecord struct {
	ID          string `storm:"id"` // primary key
	BlockHeight uint64
	TxID        string
	Reason      string
}

//NewUnscanRecord new UnscanRecord
func NewUnscanRecord(height uint64, txID, reason string) *openwallet.UnscanRecord {
	obj := openwallet.UnscanRecord{}
	obj.BlockHeight = height
	obj.TxID = txID
	obj.Reason = reason
	obj.ID = common.Bytes2Hex(crypto.SHA256([]byte(fmt.Sprintf("%d_%s", height, txID))))
	return &obj
}

type UtxoDto struct {
	TxHash   string `json:"fromHash"`
	TxIndex  int32  `json:"fromIndex"`
	Value    int64  `json:"value"`
	LockTime int64  `json:"lockTime"`
	Address  string `json:"-"`
}

func (u *UtxoDto) ScriptPubKey() string {
	scriptPubkey, err := nulsio2_addrdec.GetInputOwnerKey(u.TxHash, int64(u.TxIndex))
	if err != nil {
		fmt.Errorf("utxo.TxHash can't decode, unexpected error: %v", err)
		return ""
	}
	return hex.EncodeToString(scriptPubkey)
}

type NusBlock struct {
	Hash         string `json:"hash"`
	Height       int64  `storm:"id" json:"height"`
	Time         string `json:"time"`
	PreHash      string `json:"preHash"`
	MerkleHash   string `json:"merkleHash"`
	TxList       []*Tx  `json:"-"`
	Status       int32  `json:"status"`
	TxCount      int32  `json:"txCount"`
	Fee          int64  `json:"fee"`
	ConfirmCount int64  `json:"confirmCount"`
}

type TokenBalance struct {
	ContractAddress string `json:"contractAddress"`
	Amount          string `json:"amount"`
	Decimals        uint64 `json:"decimals"`
}

func (n *NusBlock) BlockHeader(symbol string) *openwallet.BlockHeader {
	obj := &openwallet.BlockHeader{}
	//解析json
	obj.Hash = n.Hash
	obj.Merkleroot = n.MerkleHash
	obj.Previousblockhash = n.PreHash
	obj.Height = uint64(n.Height)
	obj.Time = uint64(time.Now().Unix())
	obj.Symbol = symbol
	return obj
}


func (n *NusBlock) ToBlock() *Block {
	obj := &Block{}
	//解析json
	obj.Hash = n.Hash
	obj.Merkleroot = n.MerkleHash
	obj.Previousblockhash = n.PreHash
	obj.Height = uint32(n.Height)
	obj.Time = uint64(time.Now().Unix())
	obj.Symbol = "NULS2"
	return obj
}

type Tx struct {
	Hash         string    `json:"hash"`
	BlockHeight  int64     `json:"blockHeight"`
	Time         string    `json:"time"`
	Value        int64     `json:"value"`
	Type         int32     `json:"type"`
	Inputs       []*Input  `json:"from"`
	Outputs      []*Output `json:"to"`
	Status       int       `json:"status"`
	ConfirmCount int32     `json:"confirmCount"`
	ScriptSig    string    `json:"scriptSig"`
}

func (tx *Tx) GetTime() int64 {
	t, err := time.Parse("2006-01-02 15:04:05", tx.Time[:len(tx.Time)-4])
	if err != nil {
		return 0
	}
	timeNumber := t.UnixNano() / 1e6
	return timeNumber
}

type NulsToken struct {
	Hash            string `json:"-"`
	ContractAddress string `json:"contractAddress"`
	From            string `json:"from"`
	To              string `json:"to"`
	Value           string `json:"value"`
	Name            string `json:"name"`
	Symbol          string `json:"symbol"`
	Decimals        int64  `json:"decimals"`
}

type Output struct {
	Address       string `json:"address"`
	AssetsChainId int64  `json:"assetsChainId"`
	AssetsId      int64  `json:"assetsId"`
	LockTime      int64  `json:"lockTime"`
	Amount        string `json:"amount"`
}

type Input struct {
	Address       string `json:"address"`
	AssetsChainId int64  `json:"assetsChainId"`
	AssetsId      int64  `json:"assetsId"`
	LockTime      int64  `json:"lockTime"`
	Amount        string `json:"amount"`
}

type UnspentSort struct {
	Values     []*UtxoDto
	Comparator func(a, b *UtxoDto) int
}

type NulsBalance struct {
	Balance       decimal.Decimal
	UnLockBalance decimal.Decimal
}

func (s UnspentSort) Len() int {
	return len(s.Values)
}
func (s UnspentSort) Swap(i, j int) {
	s.Values[i], s.Values[j] = s.Values[j], s.Values[i]
}
func (s UnspentSort) Less(i, j int) bool {
	return s.Comparator(s.Values[i], s.Values[j]) < 0
}

type Nuls2Balance struct {
	Total         string `json:"total"`
	Freeze        string `json:"freeze"`
	Available     string `json:"available"`
	TimeLock      string `json:"timeLock"`
	ConsensusLock string `json:"consensusLock"`
	Nonce         string `json:"nonce"`
	NonceType     int64 `json:"nonceType"`
}
