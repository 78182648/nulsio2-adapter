package nulsio2_trans

import "encoding/hex"

type TxIn struct {
	Address       []byte
	AssetsChainId []byte
	AssetsId      []byte
	Amount        []byte
	Nonce         []byte
	Locked        []byte
}

func newTxInForEmptyTrans(vin []Vin) ([]TxIn, error) {
	var ret []TxIn

	for _, v := range vin {

		address, _ := GetBytesWithLength(AddressBase58Decode(v.Address))
		assetChainId := uint16ToLittleEndianBytes(1)
		assetsId := uint16ToLittleEndianBytes(1)

		na := WriteBigInteger(int64(v.Amount))

		nonceByte, _ := hex.DecodeString(v.Nonce)
		nonce, _ := GetBytesWithLength(nonceByte)

		lockTime := []byte{0}

		ret = append(ret, TxIn{address, assetsId, assetChainId, na, nonce, lockTime})
	}
	return ret, nil
}
