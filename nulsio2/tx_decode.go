package nulsio2

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/blocktree/go-owcrypt"
	"github.com/blocktree/nulsio2-adapter/nulsio2_trans"
	"github.com/blocktree/openwallet/v2/common"
	"github.com/blocktree/openwallet/v2/openwallet"
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
		return decoder.CreateSimpleRawNrc20Transaction(wrapper, rawTx)
		//return openwallet.Errorf(openwallet.ErrUnknownException, "nrc20 not support in nuls2.0")
	} else {
		_, err := decoder.CreateSimpleRawTransaction(wrapper, rawTx)
		return err
	}
}

//CreateSummaryRawTransaction 创建汇总交易，返回原始交易单数组
func (decoder *TransactionDecoder) CreateSummaryRawTransactionWithError(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransactionWithError, error) {
	if sumRawTx.Coin.IsContract {
		return decoder.CreateNrc20TokenSummaryRawTransaction(wrapper, sumRawTx)
	} else {
		return decoder.CreateSimpleSummaryRawTransaction(wrapper, sumRawTx)
	}
}

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateSimpleRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) (string, error) {

	var (
		decimals        = decoder.wm.Decimal()
		accountID       = rawTx.Account.AccountID
		fixFees         = big.NewInt(0)
		findAddrBalance *AddrBalance
	)

	//获取wallet
	addresses, err := wrapper.GetAddressList(0, -1, "AccountID", accountID) //wrapper.GetWallet().GetAddressesByAccount(rawTx.Account.AccountID)
	if err != nil {
		return "", err
	}

	if len(addresses) == 0 {
		return "", fmt.Errorf("[%s] have not addresses", accountID)
	}

	if len(rawTx.To) > 1 {
		return "", openwallet.Errorf(openwallet.ErrUnknownException, "nrc20 not support many address in nuls2.0")

	}

	searchAddrs := make([]string, 0)
	for _, address := range addresses {
		searchAddrs = append(searchAddrs, address.Address)
	}

	addrBalanceArray, err := decoder.wm.Blockscanner.GetBalanceByAddress(searchAddrs...)
	if err != nil {
		return "", err
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

	fixFees = common.StringNumToBigIntWithExp(decoder.wm.Config.FixFees, decimals)

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
		return "", openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAddress, "all address's balance of account is not enough")
	}

	//最后创建交易单
	hexStr, err := decoder.createSimpleRawTransaction(
		wrapper,
		rawTx,
		findAddrBalance,
		fixFees,
		"", "")
	if err != nil {
		return "", err
	}

	return hexStr, nil
}

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) createSimpleRawTransactionForNrc20Main(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction, nonce string) (string, error) {

	var (
		decimals        = decoder.wm.Decimal()
		accountID       = rawTx.Account.AccountID
		fixFees         = big.NewInt(0)
		findAddrBalance *AddrBalance
	)

	//获取wallet
	addresses, err := wrapper.GetAddressList(0, -1, "AccountID", accountID) //wrapper.GetWallet().GetAddressesByAccount(rawTx.Account.AccountID)
	if err != nil {
		return "", err
	}

	if len(addresses) == 0 {
		return "", fmt.Errorf("[%s] have not addresses", accountID)
	}

	if len(rawTx.To) > 1 {
		return "", openwallet.Errorf(openwallet.ErrUnknownException, "nrc20 not support many address in nuls2.0")

	}

	searchAddrs := make([]string, 0)
	for _, address := range addresses {
		searchAddrs = append(searchAddrs, address.Address)
	}

	addrBalanceArray, err := decoder.wm.Blockscanner.GetBalanceByAddress(searchAddrs...)
	if err != nil {
		return "", err
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

	fixFees = common.StringNumToBigIntWithExp(decoder.wm.Config.FixFees, decimals)

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
		return "", openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAddress, "all address's balance of account is not enough")
	}

	//最后创建交易单
	hexStr, err := decoder.createSimpleRawTransaction(
		wrapper,
		rawTx,
		findAddrBalance,
		fixFees,
		"", nonce)
	if err != nil {
		return "", err
	}

	return hexStr, nil
}

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateSimpleRawNrc20Transaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		decimals        = decoder.wm.Decimal()
		accountID       = rawTx.Account.AccountID
		fixFees         = big.NewInt(0)
		findAddrBalance *AddrBalance
		tokenAddress    string
		tokenDecimal    uint64
		sendAddress     string
		to              string
		totalSend       = decimal.New(0, 0)
	)

	//获取wallet
	addresses, err := wrapper.GetAddressList(0, -1, "AccountID", accountID) //wrapper.GetWallet().GetAddressesByAccount(rawTx.Account.AccountID)
	if err != nil {
		return err
	}
	if rawTx.Coin.IsContract {
		tokenAddress = rawTx.Coin.Contract.Address
		tokenDecimal = rawTx.Coin.Contract.Decimals
	} else {
		return errors.New("This is a token transaction!")
	}

	if len(rawTx.To) > 1 {
		return openwallet.Errorf(openwallet.ErrUnknownException, "nrc20 not support many toAddress")
	}

	for addr, amount := range rawTx.To {
		deamount, _ := decimal.NewFromString(amount)
		totalSend = totalSend.Add(deamount)
		to = addr
	}
	if len(addresses) == 0 {
		return fmt.Errorf("[%s] have not addresses", accountID)
	}

	searchAddrs := make([]string, 0)
	for _, address := range addresses {
		searchAddrs = append(searchAddrs, address.Address)
	}

	addrBalanceMainArray, err := decoder.wm.Blockscanner.GetBalanceByAddress(searchAddrs...)
	if err != nil {
		return err
	}

	var amountStr string
	for _, v := range rawTx.To {
		amountStr = v
		break
	}

	//地址余额从大到小排序
	sort.Slice(addrBalanceMainArray, func(i int, j int) bool {
		a_amount, _ := decimal.NewFromString(addrBalanceMainArray[i].Balance)
		b_amount, _ := decimal.NewFromString(addrBalanceMainArray[j].Balance)
		if a_amount.LessThan(b_amount) {
			return true
		} else {
			return false
		}
	})

	amount := common.StringNumToBigIntWithExp(amountStr, int32(tokenDecimal))

	fixFees = common.StringNumToBigIntWithExp(decoder.wm.Config.TokenFees, decimals)

	for _, addrBalance := range addrBalanceMainArray {

		addrBalance_BI := common.StringNumToBigIntWithExp(addrBalance.Balance, decimals)

		//总消耗数量 = 转账数量 + 手续费
		//totalAmount := new(big.Int)
		//totalAmount.Add(amount, fixFees)

		//主币余额不足查找下一个地址
		if addrBalance_BI.Cmp(fixFees) < 0 {
			continue
		}
		addrBalanceMainDecimal := common.BigIntToDecimals(fixFees, decimals)

		tokenBalance, err := decoder.wm.Api.GetTokenBalances(tokenAddress, addrBalance.Address)
		if err != nil {
			continue
		}

		addrBalanceTokenDecimal := common.BigIntToDecimals(amount, int32(tokenDecimal))

		//代币余额不足查找下一个地址
		if tokenBalance.Cmp(addrBalanceTokenDecimal) < 0 {
			continue
		}

		//只要找到一个合适使用的地址余额就停止遍历
		findAddrBalance = &AddrBalance{Address: addrBalance.Address, Balance: &addrBalanceMainDecimal, TokenBalance: &tokenBalance}
		sendAddress = addrBalance.Address
		break
	}

	if findAddrBalance == nil {
		return openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAddress, "all address's balance of account is not enough")
	}

	token := &nulsio2_trans.TxToken{
		Sender:          sendAddress,
		ContractAddress: tokenAddress,
		Value:           0,
		GasLimit:        20000,
		Price:           25,
		MethodName:      "transfer",
		ArgsCount:       2,
		Args:            []string{to, totalSend.Shift(int32(tokenDecimal)).String()},
	}
	//最后创建交易单
	errE := decoder.createSimpleNrc20RawTransaction(
		wrapper,
		rawTx,
		findAddrBalance,
		fixFees,
		"", token)
	if errE != nil {
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
		//nonce string
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

		_, createErr := decoder.createSimpleRawTransaction(
			wrapper,
			rawTx,
			&AddrBalance{Address: addrBalance.Address, Balance: &aaddrBalanceDecimal},
			feeInfo,
			"", "")
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

//CreateNrc20RawTransaction 创建合约交易

func (this *TransactionDecoder) CreateNrc20TokenSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransactionWithError, error) {

	var (
		rawTxArray         = make([]*openwallet.RawTransactionWithError, 0)
		accountID          = sumRawTx.Account.AccountID
		minTransfer        *big.Int
		retainedBalance    *big.Int
		feesSupportAccount *openwallet.AssetsAccount
		tmpNonce           string
	)

	// 如果有提供手续费账户，检查账户是否存在
	if feesAcount := sumRawTx.FeesSupportAccount; feesAcount != nil {
		account, supportErr := wrapper.GetAssetsAccountInfo(feesAcount.AccountID)
		if supportErr != nil {
			return nil, openwallet.Errorf(openwallet.ErrAccountNotFound, "can not find fees support account")
		}

		feesSupportAccount = account

		//获取手续费支持账户的地址nonce
		feesAddresses, feesSupportErr := wrapper.GetAddressList(0, 1,
			"AccountID", feesSupportAccount.AccountID)
		if feesSupportErr != nil {
			return nil, openwallet.NewError(openwallet.ErrAddressNotFound, "fees support account have not addresses")
		}

		if len(feesAddresses) == 0 {
			return nil, openwallet.Errorf(openwallet.ErrAccountNotAddress, "fees support account have not addresses")
		}

		//_, nonce, feesSupportErr := this.GetTransactionCount2(feesAddresses[0].Address)
		//if feesSupportErr != nil {
		//	return nil, openwallet.NewError(openwallet.ErrNonceInvaild, "fees support account get nonce failed")
		//}
		//tmpNonce = nonce
	}
	//tokenCoin := sumRawTx.Coin.Contract.Token
	tokenDecimals := int(sumRawTx.Coin.Contract.Decimals)
	contractAddress := sumRawTx.Coin.Contract.Address
	//coinDecimals := this.wm.Decimal()

	minTransfer, _ = ConvertFloatStringToBigInt(sumRawTx.MinTransfer, tokenDecimals)
	retainedBalance, _ = ConvertFloatStringToBigInt(sumRawTx.RetainedBalance, tokenDecimals)

	if minTransfer.Cmp(retainedBalance) < 0 {
		return nil, openwallet.Errorf(openwallet.ErrCreateRawTransactionFailed, "mini transfer amount must be greater than address retained balance")
	}

	//获取wallet
	addresses, err := wrapper.GetAddressList(sumRawTx.AddressStartIndex, sumRawTx.AddressLimit,
		"AccountID", sumRawTx.Account.AccountID)
	if err != nil {
		return nil, err
	}

	if len(addresses) == 0 {
		return nil, openwallet.Errorf(openwallet.ErrAccountNotAddress, "[%s] have not addresses", accountID)
	}

	searchAddrs := make([]string, 0)
	for _, address := range addresses {
		searchAddrs = append(searchAddrs, address.Address)
	}

	//查询Token余额
	addrBalanceArray, err := this.wm.ContractDecoder.GetTokenBalanceByAddress(sumRawTx.Coin.Contract, searchAddrs...)
	if err != nil {
		return nil, err
	}

	for _, addrBalance := range addrBalanceArray {

		//检查余额是否超过最低转账
		addrBalance_BI, _ := ConvertFloatStringToBigInt(addrBalance.Balance.Balance, tokenDecimals)

		if addrBalance_BI.Cmp(minTransfer) < 0 || addrBalance_BI.Cmp(big.NewInt(0)) <= 0 {
			continue
		}
		//计算汇总数量 = 余额 - 保留余额
		sumAmount_BI := new(big.Int)
		sumAmount_BI.Sub(addrBalance_BI, retainedBalance)

		sumAmount := common.BigIntToDecimals(sumAmount_BI, int32(tokenDecimals))
		fees, _ := decimal.NewFromString(this.wm.Config.TokenFees)

		coinBalances, createErr := this.wm.Blockscanner.GetBalanceByAddress(addrBalance.Balance.Address)
		if createErr != nil {
			continue
		}

		if len(coinBalances) <= 0 {
			continue
		}
		coinBalance, _ := decimal.NewFromString(coinBalances[0].Balance)
		//判断主币余额是否够手续费
		if coinBalance.Cmp(fees) < 0 {

			//有手续费账户支持
			if feesSupportAccount != nil {

				//通过手续费账户创建交易单
				supportAddress := addrBalance.Balance.Address
				supportAmount := decimal.Zero
				feesSupportScale, _ := decimal.NewFromString(sumRawTx.FeesSupportAccount.FeesSupportScale)
				fixSupportAmount, _ := decimal.NewFromString(sumRawTx.FeesSupportAccount.FixSupportAmount)

				//优先采用固定支持数量
				if fixSupportAmount.GreaterThan(decimal.Zero) {
					supportAmount = fixSupportAmount
				} else {
					//没有固定支持数量，有手续费倍率，计算支持数量
					if feesSupportScale.GreaterThan(decimal.Zero) {
						supportAmount = feesSupportScale.Mul(fees)
					} else {
						//默认支持数量为手续费
						supportAmount = fees
					}
				}

				this.wm.Log.Debugf("create transaction for fees support account")
				this.wm.Log.Debugf("fees account: %s", feesSupportAccount.AccountID)
				this.wm.Log.Debugf("mini support amount: %s", fees.String())
				this.wm.Log.Debugf("allow support amount: %s", supportAmount.String())
				this.wm.Log.Debugf("support address: %s", supportAddress)

				supportCoin := openwallet.Coin{
					Symbol:     sumRawTx.Coin.Symbol,
					IsContract: false,
				}

				//创建一笔交易单
				rawTx := &openwallet.RawTransaction{
					Coin:    supportCoin,
					Account: feesSupportAccount,
					To: map[string]string{
						addrBalance.Balance.Address: supportAmount.String(),
					},
					Required: 1,
				}

				hexStr, createTxErr := this.createSimpleRawTransactionForNrc20Main(wrapper, rawTx, tmpNonce)

				rawTxWithErr := &openwallet.RawTransactionWithError{
					RawTx: rawTx,
					Error: openwallet.ConvertError(createTxErr),
				}

				//创建成功，添加到队列
				rawTxArray = append(rawTxArray, rawTxWithErr)

				//取哈希后8位
				tmpNonce = hexStr[len(hexStr)-16:]

				//汇总下一个
				continue
			}
		}

		this.wm.Log.Debugf("balance: %v", addrBalance.Balance.Balance)
		this.wm.Log.Debugf("%s fees: %v", sumRawTx.Coin.Symbol, fees)
		this.wm.Log.Debugf("sumAmount: %v", sumAmount)

		//创建一笔交易单
		rawTx := &openwallet.RawTransaction{
			Coin:    sumRawTx.Coin,
			Account: sumRawTx.Account,
			To: map[string]string{
				sumRawTx.SummaryAddress: sumAmount.StringFixed(int32(tokenDecimals)),
			},
			Required: 1,
		}
		token := &nulsio2_trans.TxToken{
			Sender:          addrBalance.Balance.Address,
			ContractAddress: contractAddress,
			Value:           0,
			GasLimit:        20000,
			Price:           25,
			MethodName:      "transfer",
			ArgsCount:       2,
			Args:            []string{sumRawTx.SummaryAddress, sumAmount_BI.String()},
		}
		fixFees := common.StringNumToBigIntWithExp(this.wm.Config.TokenFees, this.wm.Decimal())
		//最后创建交易单
		createTxErr := this.createSimpleNrc20RawTransaction(
			wrapper,
			rawTx,
			&AddrBalance{Address: addrBalance.Balance.Address, Balance: &coinBalance, TokenBalance: &sumAmount},
			fixFees,
			"", token)
		rawTxWithErr := &openwallet.RawTransactionWithError{
			RawTx: rawTx,
			Error: createTxErr,
		}

		//创建成功，添加到队列
		rawTxArray = append(rawTxArray, rawTxWithErr)

	}

	return rawTxArray, nil
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
			signature, _, sigErr := owcrypt.Signature(keyBytes, nil, txHash, owcrypt.ECC_CURVE_SECP256K1)
			if sigErr != owcrypt.SUCCESS {
				return fmt.Errorf("transaction hash sign failed")
			}
			//signature = append(signature, v)
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

	fees := rawTx.Fees

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
	callData, nonce string) (string, error) {

	var (
		err              error
		vins             = make([]nulsio2_trans.Vin, 0)
		vouts            = make([]nulsio2_trans.Vout, 0)
		accountTotalSent = decimal.Zero
		txFrom           = make([]string, 0)
		txTo             = make([]string, 0)
	)

	if addrBalance == nil {
		return "", fmt.Errorf("Receiver addresses is empty! ")
	}

	fromAddress, err := decoder.wm.Api.GetAddressBalance(addrBalance.Address, 1)
	if err != nil {
		return "", openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAddress, "can't find the address:"+err.Error())
	}
	//fromAddress.Nonce = "0000000000000000"
	if nonce != "" {
		fromAddress.Nonce = nonce
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

		//累加
		accountTotalSent = accountTotalSent.Add(amountDecimal)

		amountDecimal = amountDecimal.Shift(decoder.wm.Decimal())

		to, err := decoder.wm.Api.GetAddressBalance(toAddress, 1)
		if err != nil {
			return "", openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAddress, "can't find the address:"+err.Error())
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
		return "", fmt.Errorf("create transaction failed, unexpected error: %v", err)
	}

	rawTx.RawHex = signTrans

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]*openwallet.KeySignature)
	}

	//装配签名
	keySigs := make([]*openwallet.KeySignature, 0)

	addr, err := wrapper.GetAddress(addrBalance.Address)
	if err != nil {
		return "", err
	}

	beSignHexHex, err := hex.DecodeString(signTrans)
	if err != nil {
		return "", err
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

	rawTx.Fees = "0.001" //默认手续费

	feesDec, _ := decimal.NewFromString(rawTx.Fees)
	accountTotalSent = accountTotalSent.Add(feesDec)
	accountTotalSent = decimal.Zero.Sub(accountTotalSent)

	rawTx.Signatures[rawTx.Account.AccountID] = keySigs
	rawTx.IsBuilt = true
	rawTx.TxAmount = accountTotalSent.StringFixed(decoder.wm.Decimal())
	rawTx.TxFrom = txFrom
	rawTx.TxTo = txTo

	return messageStr, nil
}

//createSimpleRawTransaction 创建原始交易单
func (decoder *TransactionDecoder) createSimpleNrc20RawTransaction(
	wrapper openwallet.WalletDAI,
	rawTx *openwallet.RawTransaction,
	addrBalance *AddrBalance,
	feeInfo *big.Int,
	callData string,
	token *nulsio2_trans.TxToken) *openwallet.Error {

	var (
		err              error
		vins             = make([]nulsio2_trans.Vin, 0)
		vouts            = make([]nulsio2_trans.Vout, 0)
		accountTotalSent = decimal.Zero
		txFrom           = make([]string, 0)
		txTo             = make([]string, 0)
	)

	if addrBalance == nil {
		return openwallet.Errorf(openwallet.ErrUnknownException, "Receiver addresses is empty! ")
	}

	fromAddress, err := decoder.wm.Api.GetAddressBalance(addrBalance.Address, 1)
	if err != nil {
		return openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAddress, "can't find the address:"+err.Error())
	}

	feeDe, _ := decimal.NewFromString(decoder.wm.Config.TokenFees)
	if feeDe.Equal(decimal.Zero) {
		feeDe, _ = decimal.NewFromString("0.01")
	}
	//装配输入(直接使用手续费)
	in := nulsio2_trans.Vin{
		Address:       addrBalance.Address,
		Nonce:         fromAddress.Nonce,
		AssetsChainId: 1,
		AssetsId:      1,
		Amount:        uint64(feeDe.Shift(decoder.wm.Decimal()).IntPart()),
	}
	vins = append(vins, in)
	txFrom = append(txFrom, fmt.Sprintf("%s:%s", addrBalance.Address, addrBalance.TokenBalance.String()))

	//装配输出
	for toAddress, amount := range rawTx.To {

		//to, err := decoder.wm.Api.GetAddressBalance(toAddress, 1)
		//if err != nil {
		//	return openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAddress, "can't find the address:"+err.Error())
		//}
		//out := nulsio2_trans.Vout{
		//	Address:       toAddress,
		//	Nonce:         to.Nonce,
		//	AssetsChainId: 1,
		//	AssetsId:      1,
		//	Amount:        uint64(addrBalance.Balance.Shift(decoder.wm.Decimal()).IntPart()),
		//}
		//vouts = append(vouts, out)

		txTo = append(txTo, fmt.Sprintf("%s:%s", toAddress, amount))
	}

	//锁定时间
	lockTime := uint32(0)

	//追加手续费支持
	replaceable := false

	/////////构建空交易单
	signTrans, _, err := nulsio2_trans.CreateEmptyRawTransaction(vins, vouts, "", lockTime, replaceable, token)

	if err != nil {
		return openwallet.Errorf(openwallet.ErrUnknownException, "create transaction failed, unexpected error: %v"+err.Error())
	}

	rawTx.RawHex = signTrans

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]*openwallet.KeySignature)
	}

	//装配签名
	keySigs := make([]*openwallet.KeySignature, 0)

	addr, err := wrapper.GetAddress(addrBalance.Address)
	if err != nil {
		return openwallet.Errorf(openwallet.ErrUnknownException, err.Error())
	}

	beSignHexHex, err := hex.DecodeString(signTrans)
	if err != nil {
		return openwallet.Errorf(openwallet.ErrUnknownException, err.Error())
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

	rawTx.Fees = decoder.wm.Config.TokenFees //默认手续费

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
