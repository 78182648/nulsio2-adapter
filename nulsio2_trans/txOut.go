package nulsio2_trans

type TxOut struct {
	Address       []byte
	AssetsChainId []byte
	AssetsId      []byte
	Amount        []byte
	Locked        []byte
}

func newTxOutForEmptyTrans(vout []Vout) ([]TxOut, error) {
	var ret []TxOut

	for _, v := range vout {
		address, _ := GetBytesWithLength(AddressBase58Decode(v.Address))
		assetChainId := uint16ToLittleEndianBytes(1)
		assetsId := uint16ToLittleEndianBytes(1)

		na := WriteBigInteger(int64(v.Amount))

		//nonceByte, _ := hex.DecodeString(v.Nonce)
		//nonce, _ := GetBytesWithLength(nonceByte)

		lockTime := []byte{0,0,0,0,0,0,0,0,0}

		ret = append(ret, TxOut{address, assetsId, assetChainId, na, lockTime})
	}
	return ret, nil
}
