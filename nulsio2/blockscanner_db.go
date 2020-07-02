package nulsio2

import (
	"fmt"
	"github.com/blocktree/openwallet/v2/openwallet"
)

//SaveLocalBlockHead 记录区块高度和hash到本地
func (bs *NULSBlockScanner) SaveLocalBlockHead(blockHeight uint32, blockHash string) error {

	if bs.BlockchainDAI == nil {
		return fmt.Errorf("Blockchain DAI is not setup ")
	}

	header := &openwallet.BlockHeader{
		Hash:   blockHash,
		Height: uint64(blockHeight),
		Fork:   false,
		Symbol: bs.wm.Symbol(),
	}

	return bs.BlockchainDAI.SaveCurrentBlockHead(header)
}

//GetLocalBlockHead 获取本地记录的区块高度和hash
func (bs *NULSBlockScanner) GetLocalBlockHead() (uint64, string, error) {

	if bs.BlockchainDAI == nil {
		return 0, "", fmt.Errorf("Blockchain DAI is not setup ")
	}

	header, err := bs.BlockchainDAI.GetCurrentBlockHead(bs.wm.Symbol())
	if err != nil {
		return 0, "", err
	}

	return uint64(header.Height), header.Hash, nil
}

//SaveLocalBlock 记录本地新区块
func (bs *NULSBlockScanner) SaveLocalBlock(blockHeader *Block) error {

	if bs.BlockchainDAI == nil {
		return fmt.Errorf("Blockchain DAI is not setup ")
	}

	header := &openwallet.BlockHeader{
		Hash:              blockHeader.Hash,
		Merkleroot:        blockHeader.Merkleroot,
		Previousblockhash: blockHeader.Previousblockhash,
		Height:            uint64(blockHeader.Height),
		Time:              uint64(blockHeader.Time),
		Symbol:            bs.wm.Symbol(),
	}

	return bs.BlockchainDAI.SaveLocalBlockHead(header)
}

//GetLocalBlock 获取本地区块数据
func (bs *NULSBlockScanner) GetLocalBlock(height uint32) (*Block, error) {

	if bs.BlockchainDAI == nil {
		return nil, fmt.Errorf("Blockchain DAI is not setup ")
	}

	header, err := bs.BlockchainDAI.GetLocalBlockHeadByHeight(uint64(height), bs.wm.Symbol())
	if err != nil {
		return nil, err
	}

	block := &Block{
		Height: uint32(header.Height),
	}
	block.Hash = header.Hash
	block.Previousblockhash = header.Previousblockhash
	return block, nil
}

//SaveUnscanRecord 保存交易记录到钱包数据库
func (bs *NULSBlockScanner) SaveUnscanRecord(record *openwallet.UnscanRecord) error {

	if bs.BlockchainDAI == nil {
		return fmt.Errorf("Blockchain DAI is not setup ")
	}

	return bs.BlockchainDAI.SaveUnscanRecord(record)
}

//DeleteUnscanRecord 删除指定高度的未扫记录
func (bs *NULSBlockScanner) DeleteUnscanRecord(height uint32) error {

	if bs.BlockchainDAI == nil {
		return fmt.Errorf("Blockchain DAI is not setup ")
	}

	return bs.BlockchainDAI.DeleteUnscanRecordByHeight(uint64(height), bs.wm.Symbol())
}

func (bs *NULSBlockScanner) GetUnscanRecords() ([]*openwallet.UnscanRecord, error) {

	if bs.BlockchainDAI == nil {
		return nil, fmt.Errorf("Blockchain DAI is not setup ")
	}

	return bs.BlockchainDAI.GetUnscanRecords(bs.wm.Symbol())
}
