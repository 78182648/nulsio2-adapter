package nulsio2_trans

type TxToken struct {
	Sender          string
	ContractAddress string
	Value           uint64
	GasLimit        uint64
	Price           uint64
	MethodName      string
	ArgsCount       int64
	Args            []string
}

func newTxTokenToBytes(tx *TxToken) ([]byte, error) {
	ret := make([]byte, 0)
	sendBytes:= AddressBase58Decode(tx.Sender)
	ret = append(ret, sendBytes...)
	//contractAddress := nulsio2_addrdec.Base58Decode([]byte(tx.ContractAddress))
	contractAddress := AddressBase58Decode(tx.ContractAddress)
	ret = append(ret, contractAddress...)
	valueBytes := WriteBigInteger(int64(tx.Value))
	ret = append(ret, valueBytes...)
	gasLimitBytes := int64ToLittleEndianBytes(tx.GasLimit)
	ret = append(ret, gasLimitBytes...)
	price := int64ToLittleEndianBytes(tx.Price)
	ret = append(ret, price...)
	methodName, _ := GetBytesWithLength([]byte(tx.MethodName))
	ret = append(ret, methodName...)
	ret = append(ret, 0)
	ret = append(ret, byte(tx.ArgsCount))
	for _, v := range tx.Args {
		arg, _ := GetBytesWithLength([]byte(v))
		ret = append(ret, 1)
		ret = append(ret, arg...)
	}
	return ret, nil
}
