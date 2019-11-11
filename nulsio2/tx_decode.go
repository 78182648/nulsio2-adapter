package nulsio2

import (
	"encoding/hex"
	"fmt"
	"github.com/blocktree/go-owcrypt"
	"github.com/blocktree/nulsio2-adapter/nulsio2_trans"
	"github.com/blocktree/openwallet/common"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/shopspring/decimal"
	"math/big"
	"sort"
	"time"
)

var (
	DEFAULT_GAS_LIMIT = "250000"
	DEFAULT_GAS_PRICE = decimal.New(4, -7)
)

type TransactionDecoder struct {
	openwallet.TransactionDecoderBase
	wm *WalletManager //钱包管理者
}

//NewTransactionDecoder 交易单解析器
func NewTransactionDecoder(wm *WalletManager) *TransactionDecoder {
	decoder := TransactionDecoder{}
	decoder.wm = wm
	return &decoder
}

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	if rawTx.Coin.IsContract {
		return openwallet.Errorf(openwallet.ErrUnknownException, "nrc20 not support in nuls2.0")
	} else {
		return decoder.CreateSimpleRawTransaction(wrapper, rawTx)
	}
}

//CreateSummaryRawTransaction 创建汇总交易，返回原始交易单数组
func (decoder *TransactionDecoder) CreateSummaryRawTransactionWithError(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransactionWithError, error) {
	if sumRawTx.Coin.IsContract {
		return nil, openwallet.Errorf(openwallet.ErrUnknownException, "nrc20 not support in nuls2.0")
	} else {
		return decoder.CreateSimpleSummaryRawTransaction(wrapper, sumRawTx)
	}
}

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateSimpleRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		decimals        = decoder.wm.Decimal()
		accountID       = rawTx.Account.AccountID
		fixFees         = big.NewInt(0)
		findAddrBalance *AddrBalance
	)

	//获取wallet
	addresses, err := wrapper.GetAddressList(0, -1, "AccountID", accountID) //wrapper.GetWallet().GetAddressesByAccount(rawTx.Account.AccountID)
	if err != nil {
		return err
	}

	if len(addresses) == 0 {
		return fmt.Errorf("[%s] have not addresses", accountID)
	}

	searchAddrs := make([]string, 0)
	for _, address := range addresses {
		searchAddrs = append(searchAddrs, address.Address)
	}

	addrBalanceArray, err := decoder.wm.Blockscanner.GetBalanceByAddress(searchAddrs...)
	if err != nil {
		return err
	}

	var amountStr string
	for _, v := range rawTx.To {
		amountStr = v
		break
	}

	//地址余额从大到小排序
	sort.Slice(addrBalanceArray, func(i int, j int) bool {
		a_amount, _ := decimal.NewFromString(addrBalanceArray[i].Balance)
		b_amount, _ := decimal.NewFromString(addrBalanceArray[j].Balance)
		if a_amount.LessThan(b_amount) {
			return true
		} else {
			return false
		}
	})

	amount := common.StringNumToBigIntWithExp(amountStr, decimals)

	if len(rawTx.FeeRate) > 0 {
		fixFees = common.StringNumToBigIntWithExp(rawTx.FeeRate, decimals)
	} else {
		fixFees = common.StringNumToBigIntWithExp(decoder.wm.Config.FixFees, decimals)
	}

	for _, addrBalance := range addrBalanceArray {

		addrBalance_BI := common.StringNumToBigIntWithExp(addrBalance.Balance, decimals)

		//总消耗数量 = 转账数量 + 手续费
		totalAmount := new(big.Int)
		totalAmount.Add(amount, fixFees)

		//余额不足查找下一个地址
		if addrBalance_BI.Cmp(totalAmount) < 0 {
			continue
		}

		addrBalanceDecimal := common.BigIntToDecimals(totalAmount, decimals)

		//只要找到一个合适使用的地址余额就停止遍历
		findAddrBalance = &AddrBalance{Address: addrBalance.Address, Balance: &addrBalanceDecimal}
		break
	}

	if findAddrBalance == nil {
		return openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAddress, "all address's balance of account is not enough")
	}

	//最后创建交易单
	err = decoder.createSimpleRawTransaction(
		wrapper,
		rawTx,
		findAddrBalance,
		fixFees,
		"")
	if err != nil {
		return err
	}

	return nil
}

//CreateSimpleSummaryRawTransaction 创建主币汇总交易
func (decoder *TransactionDecoder) CreateSimpleSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransactionWithError, error) {

	var (
		decimals        = decoder.wm.Decimal()
		rawTxArray      = make([]*openwallet.RawTransaction, 0)
		accountID       = sumRawTx.Account.AccountID
		minTransfer     = common.StringNumToBigIntWithExp(sumRawTx.MinTransfer, decimals)
		retainedBalance = common.StringNumToBigIntWithExp(sumRawTx.RetainedBalance, decimals)
		fixFees         = big.NewInt(0)
		feeInfo         *big.Int
	)

	if minTransfer.Cmp(retainedBalance) < 0 {
		return nil, fmt.Errorf("mini transfer amount must be greater than address retained balance")
	}

	//获取wallet
	addresses, err := wrapper.GetAddressList(sumRawTx.AddressStartIndex, sumRawTx.AddressLimit,
		"AccountID", sumRawTx.Account.AccountID)
	if err != nil {
		return nil, err
	}

	if len(addresses) == 0 {
		return nil, fmt.Errorf("[%s] have not addresses", accountID)
	}

	searchAddrs := make([]string, 0)
	for _, address := range addresses {
		searchAddrs = append(searchAddrs, address.Address)
	}

	addrBalanceArray, err := decoder.wm.Blockscanner.GetBalanceByAddress(searchAddrs...)
	if err != nil {
		return nil, err
	}

	if len(sumRawTx.FeeRate) > 0 {
		fixFees = common.StringNumToBigIntWithExp(sumRawTx.FeeRate, decimals)
	} else {
		fixFees = common.StringNumToBigIntWithExp(decoder.wm.Config.FixFees, decimals)
	}

	for _, addrBalance := range addrBalanceArray {

		//检查余额是否超过最低转账
		addrBalance_BI := common.StringNumToBigIntWithExp(addrBalance.Balance, decimals)

		if addrBalance_BI.Cmp(minTransfer) < 0 || addrBalance_BI.Cmp(big.NewInt(0)) <= 0 {
			continue
		}
		//计算汇总数量 = 余额 - 保留余额
		sumAmount_BI := new(big.Int)
		sumAmount_BI.Sub(addrBalance_BI, retainedBalance)

		//减去手续费
		sumAmount_BI.Sub(sumAmount_BI, fixFees)
		if sumAmount_BI.Cmp(big.NewInt(0)) <= 0 {
			continue
		}

		sumAmount := common.BigIntToDecimals(sumAmount_BI, decimals)
		feesAmount := common.BigIntToDecimals(fixFees, decimals)

		decoder.wm.Log.Debugf("balance: %v", addrBalance.Balance)
		decoder.wm.Log.Debugf("fees: %v", feesAmount)
		decoder.wm.Log.Debugf("sumAmount: %v", sumAmount)

		//创建一笔交易单
		rawTx := &openwallet.RawTransaction{
			Coin:    sumRawTx.Coin,
			Account: sumRawTx.Account,
			To: map[string]string{
				sumRawTx.SummaryAddress: sumAmount.StringFixed(decoder.wm.Decimal()),
			},
			Required: 1,
		}

		aaddrBalanceDecimal := common.BigIntToDecimals(addrBalance_BI, decimals)

		createErr := decoder.createSimpleRawTransaction(
			wrapper,
			rawTx,
			&AddrBalance{Address: addrBalance.Address, Balance: &aaddrBalanceDecimal},
			feeInfo,
			"")
		if createErr != nil {
			return nil, createErr
		}

		//创建成功，添加到队列
		rawTxArray = append(rawTxArray, rawTx)

	}

	raTxWithErr := make([]*openwallet.RawTransactionWithError, 0)

	for _, tx := range rawTxArray {
		raTxWithErr = append(raTxWithErr, &openwallet.RawTransactionWithError{
			RawTx: tx,
			Error: nil,
		})
	}
	return raTxWithErr, nil
}

//SignRawTransaction 签名交易单
func (decoder *TransactionDecoder) SignRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	if rawTx.Signatures == nil || len(rawTx.Signatures) == 0 {
		//this.wm.Log.Std.Error("len of signatures error. ")
		return fmt.Errorf("transaction signature is empty")
	}

	key, err := wrapper.HDKey()
	if err != nil {
		return err
	}

	keySignatures := rawTx.Signatures[rawTx.Account.AccountID]
	if keySignatures != nil {
		for _, keySignature := range keySignatures {

			childKey, err := key.DerivedKeyWithPath(keySignature.Address.HDPath, keySignature.EccType)
			keyBytes, err := childKey.GetPrivateKeyBytes()
			txHash, err := hex.DecodeString(keySignature.Message)
			if err != nil {
				return err
			}

			//签名交易（无特殊签名）
			signature, retCode := owcrypt.Signature(keyBytes, nil, 0, txHash, 32, owcrypt.ECC_CURVE_SECP256K1)
			if retCode != owcrypt.SUCCESS {
				return fmt.Errorf("transaction hash sign failed, unexpected error: Failed to sign message!")
			}


			keySignature.Signature = hex.EncodeToString(signature)
		}
	}

	rawTx.Signatures[rawTx.Account.AccountID] = keySignatures

	return nil
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	rawHex, err := hex.DecodeString(rawTx.RawHex)
	if err != nil {
		return err
	}

	if rawTx.Signatures == nil || len(rawTx.Signatures) == 0 {
		//this.wm.Log.Std.Error("len of signatures error. ")
		return fmt.Errorf("transaction signature is empty")
	}

	sigPubByte := make([]byte, 0)

	for accountID, keySignatures := range rawTx.Signatures {
		decoder.wm.Log.Debug("accountID Signatures:", accountID)
		for _, keySignature := range keySignatures {

			signature, _ := hex.DecodeString(keySignature.Signature)
			pub, _ := hex.DecodeString(keySignature.Address.PublicKey)
			//pub = owcrypt.PointCompress(pub, owcrypt.ECC_CURVE_SECP256K1)

			sigPub := &nulsio2_trans.SigPub{
				pub,
				signature,
			}

			result := make([]byte, 0)
			result = append(result, byte(len(pub)))
			result = append(result, pub...)

			//result = append(result, 0)
			resultSig := make([]byte, 0)
			resultSig = append(resultSig, sigPub.ToBytes()...)

			result = append(result, resultSig...)

			sigPubByte = append(sigPubByte, result...)
		}
	}

	sigPubByte, _ = nulsio2_trans.GetBytesWithLength(sigPubByte)

	rawBytes := make([]byte, 0)
	rawBytes = append(rawBytes, rawHex...)
	//rawBytes = rawBytes[:len(rawBytes)-1]
	rawBytes = append(rawBytes, sigPubByte...)
	rawTx.IsCompleted = true
	rawTx.RawHex = hex.EncodeToString(rawBytes)

	_, err = decoder.wm.Api.VaildTransaction(rawTx.RawHex)
	if err != nil {
		return err
	}

	return nil
}

//SendRawTransaction 广播交易单
func (decoder *TransactionDecoder) SubmitRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {

	//先加载是否有配置文件
	//err := decoder.wm.loadConfig()
	//if err != nil {
	//	return err
	//}

	if len(rawTx.RawHex) == 0 {
		return nil, fmt.Errorf("transaction hex is empty")
	}

	if !rawTx.IsCompleted {
		return nil, fmt.Errorf("transaction is not completed validation")
	}

	txId, err := decoder.wm.Api.SendRawTransaction(rawTx.RawHex)
	if err != nil {
		return nil, err
	}
	rawTx.TxID = txId

	decimals := int32(0)
	//fees := "0"
	//if rawTx.Coin.IsContract {
	//	decimals = int32(rawTx.Coin.Contract.Decimals)
	//	fees = "0"
	//} else {
	//	decimals = int32(decoder.wm.Decimal())
	fees := rawTx.Fees
	//}

	//rawTx.TxID = txid
	rawTx.IsSubmit = true

	//记录一个交易单
	tx := &openwallet.Transaction{
		From:       rawTx.TxFrom,
		To:         rawTx.TxTo,
		Amount:     rawTx.TxAmount,
		Coin:       rawTx.Coin,
		TxID:       rawTx.TxID,
		Decimal:    decimals,
		AccountID:  rawTx.Account.AccountID,
		Fees:       fees,
		SubmitTime: time.Now().Unix(),
	}

	tx.WxID = openwallet.GenTransactionWxID(tx)

	return tx, nil
}

//GetRawTransactionFeeRate 获取交易单的费率
func (decoder *TransactionDecoder) GetRawTransactionFeeRate() (feeRate string, unit string, err error) {
	rate, err := decoder.wm.EstimateFeeRate()
	if err != nil {
		return "", "", err
	}

	return rate.StringFixed(decoder.wm.Decimal()), "K", nil
}

//createSimpleRawTransaction 创建原始交易单
func (decoder *TransactionDecoder) createSimpleRawTransaction(
	wrapper openwallet.WalletDAI,
	rawTx *openwallet.RawTransaction,
	addrBalance *AddrBalance,
	feeInfo *big.Int,
	callData string) error {

	var (
		err              error
		vins             = make([]nulsio2_trans.Vin, 0)
		vouts            = make([]nulsio2_trans.Vout, 0)
		accountTotalSent = decimal.Zero
		txFrom           = make([]string, 0)
		txTo             = make([]string, 0)
	)

	if addrBalance == nil {
		return fmt.Errorf("Receiver addresses is empty! ")
	}

	fromAddress, err := decoder.wm.Api.GetAddressBalance(addrBalance.Address, 1)
	if err != nil {
		return openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAddress, "can't find the address:"+err.Error())
	}

	//装配输入
	in := nulsio2_trans.Vin{
		Address:       addrBalance.Address,
		Nonce:         fromAddress.Nonce,
		AssetsChainId: 1,
		AssetsId:      1,
		Amount:        uint64(addrBalance.Balance.Shift(decoder.wm.Decimal()).IntPart()),
	}
	vins = append(vins, in)
	txFrom = append(txFrom, fmt.Sprintf("%s:%s", addrBalance.Address, addrBalance.Balance.String()))

	//装配输入
	for toAddress, amount := range rawTx.To {

		amountDecimal, _ := decimal.NewFromString(amount)
		amountDecimal = amountDecimal.Shift(decoder.wm.Decimal())

		to, err := decoder.wm.Api.GetAddressBalance(toAddress, 1)
		if err != nil {
			return openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAddress, "can't find the address:"+err.Error())
		}
		out := nulsio2_trans.Vout{
			Address:       toAddress,
			Nonce:         to.Nonce,
			AssetsChainId: 1,
			AssetsId:      1,
			Amount:        uint64(amountDecimal.IntPart()),
		}
		vouts = append(vouts, out)

		txTo = append(txTo, fmt.Sprintf("%s:%s", toAddress, amount))
	}

	//锁定时间
	lockTime := uint32(0)

	//追加手续费支持
	replaceable := false

	/////////构建空交易单
	signTrans, _, err := nulsio2_trans.CreateEmptyRawTransaction(vins, vouts, "", lockTime, replaceable, nil)

	if err != nil {
		return fmt.Errorf("create transaction failed, unexpected error: %v", err)
	}

	rawTx.RawHex = signTrans

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]*openwallet.KeySignature)
	}

	//装配签名
	keySigs := make([]*openwallet.KeySignature, 0)


	addr, err := wrapper.GetAddress(addrBalance.Address)
	if err != nil {
		return err
	}

	beSignHexHex, err := hex.DecodeString(signTrans)
	if err != nil {
		return err
	}

	message := nulsio2_trans.Sha256Twice(beSignHexHex) //sha256

	messageStr := hex.EncodeToString(message)


	signature := openwallet.KeySignature{
		EccType: decoder.wm.Config.CurveType,
		Nonce:   "",
		Address: addr,
		Message: messageStr,
	}

	keySigs = append(keySigs, &signature)

	feesDec, _ := decimal.NewFromString(rawTx.Fees)
	accountTotalSent = accountTotalSent.Add(feesDec)
	accountTotalSent = decimal.Zero.Sub(accountTotalSent)

	rawTx.Signatures[rawTx.Account.AccountID] = keySigs
	rawTx.IsBuilt = true
	rawTx.TxAmount = accountTotalSent.StringFixed(decoder.wm.Decimal())
	rawTx.TxFrom = txFrom
	rawTx.TxTo = txTo

	return nil
}

func appendOutput(output map[string]decimal.Decimal, address string, amount decimal.Decimal) map[string]decimal.Decimal {
	if origin, ok := output[address]; ok {
		origin = origin.Add(amount)
		output[address] = origin
	} else {
		output[address] = amount
	}
	return output
}
