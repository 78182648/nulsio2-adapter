package nulsio2_trans

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/blocktree/go-owcrypt"
)

type Vin struct {
	Address  string
	AssetsChainId uint64
	AssetsId uint64
	Amount   uint64
	Nonce   string
	LockTime uint64
}


type Vout struct {
	Address  string
	AssetsChainId uint64
	AssetsId uint64
	Amount   uint64
	Nonce   string
	LockTime uint64
}

type TxUnlock struct {
	PrivateKey   []byte
	LockScript   string
	RedeemScript string
	Amount       uint64
	Address      string
}

type Token struct {
	Sender          string
	ContractAddress string
	Value           int64
	GasLimit        int64
	MethodName      string
	ArgsCount       int64
	Args            []interface{}
}

const (
	DefaultTxVersion = uint32(2)
	DefaultHashType  = uint32(1)
)

func CreateEmptyRawTransaction(vins []Vin, vouts []Vout, remark string, lockTime uint32, replaceable bool,txData *TxToken) (string, []byte, error) {
	emptyTrans, err := newTransaction(vins, vouts, nil, lockTime, txData,replaceable)
	if err != nil {
		return "", nil, err
	}

	txBytes, err := emptyTrans.encodeToBytes()
	if err != nil {
		return "", nil, err
	}

	result, _ := json.Marshal(emptyTrans)

	return hex.EncodeToString(txBytes), result, nil
}

func SignTransactionMessage(message []byte, prikey []byte) ([]byte, error) {

	signature, retCode := owcrypt.Signature(prikey, nil, 0, message, 32, owcrypt.ECC_CURVE_SECP256K1)
	if retCode != owcrypt.SUCCESS {
		return nil, errors.New("Failed to sign message!")
	}

	return signature,nil

}

type SigPub struct {
	PublicKey []byte
	Signature []byte
}



func (sp SigPub) ToBytes() []byte {
	r := sp.Signature[:32]
	s := sp.Signature[32:]
	if r[0]&0x80 == 0x80 {
		r = append([]byte{0x00}, r...)
	} else {
		for i := 0; i < 32; i++ {
			if r[i] == 0 && r[i+1]&0x80 != 0x80 {
				r = r[1:]
			} else {
				break
			}
		}
	}
	if s[0]&0x80 == 0x80 {
		s = append([]byte{0}, s...)
	} else {
		for i := 0; i < 32; i++ {
			if s[i] == 0 && s[i+1]&0x80 != 0x80 {
				s = s[1:]
			} else {
				break
			}
		}
	}

	r = append([]byte{byte(len(r))}, r...)
	r = append([]byte{0x02}, r...)
	s = append([]byte{byte(len(s))}, s...)
	s = append([]byte{0x02}, s...)

	rs := append(r, s...)
	rs = append([]byte{byte(len(rs))}, rs...)
	rs = append([]byte{0x30}, rs...)
	rs = append([]byte{byte(len(rs))}, rs...)

	return rs
}



func CreateRawTransactionHashForSig(txHex string, unlockData []TxUnlock) ([]string, error) {


	return nil, nil
}

