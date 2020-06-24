package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/account"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/transactions"
	"strings"
)

// 修改交易池内的手续费报价
func (api *DeprecatedApiService) quoteFee(params map[string]string) map[string]string {

	result := make(map[string]string)
	trsid, ok1 := params["txhash"]
	if !ok1 {
		result["err"] = "Param txhash must."
		return result
	}
	var trshx []byte
	if txhx, e := hex.DecodeString(trsid); e == nil && len(txhx) == 32 {
		trshx = txhx
	} else {
		result["err"] = "Transaction hash error."
		return result
	}
	// 查询交易
	tx, ok := api.txpool.CheckTxExistByHash(trshx)
	if !ok {
		result["err"] = "Not find transaction in txpool."
		return result
	}
	// fee
	feestr, ok2 := params["fee"]
	if !ok2 {
		result["err"] = "Param fee must."
		return result
	}
	feeamt, e2 := fields.NewAmountFromFinString(feestr)
	if e2 != nil {
		result["err"] = "Param fee format error."
		return result
	}
	// change fee
	tx.SetFee(feeamt)
	// password
	password_or_privatekey, ok3 := params["password"]
	if !ok3 {
		result["err"] = "param password must."
		return result
	}
	var acc *account.Account = nil
	privatekey, e2 := hex.DecodeString(password_or_privatekey)
	if len(password_or_privatekey) == 64 && e2 == nil {
		acc, e2 = account.GetAccountByPriviteKey(privatekey)
		if e2 != nil {
			result["err"] = "Privite Key Error"
			return result
		}
	} else {
		acc = account.CreateAccountByPassword(password_or_privatekey)
	}
	// check
	if fields.Address(acc.Address).Equal(tx.GetAddress()) != true {
		result["err"] = "Tx fee address password error."
		return result
	}
	// 私钥
	allPrivateKeyBytes := make(map[string][]byte, 1)
	allPrivateKeyBytes[string(acc.Address)] = acc.PrivateKey
	// do sign
	tx.FillNeedSigns(allPrivateKeyBytes, nil)
	// add to pool
	err4 := api.txpool.AddTx(tx)
	if err4 != nil {
		result["err"] = err4.Error()
		return result
	}
	// ok
	result["status"] = "ok"
	return result

}

// 通过 hx 获取交易简介
func (api *DeprecatedApiService) getTransactionIntro(params map[string]string) map[string]string {
	result := make(map[string]string)
	trsid, ok1 := params["id"]
	if !ok1 {
		result["err"] = "param id must."
		return result
	}
	var trshx []byte
	if txhx, e := hex.DecodeString(trsid); e == nil && len(txhx) == 32 {
		trshx = txhx
	} else {
		result["err"] = "transaction hash error."
		return result
	}
	// 查询交易

	store := api.blockchain.State().BlockStore()

	blkhei, trsresbytes, err := store.ReadTransactionBytesByHash(trshx)
	if err != nil {
		result["err"] = err.Error()
		return result
	}
	if trsresbytes == nil {
		result["err"] = "transaction not fond."
		return result
	}

	trsres, _, err := transactions.ParseTransaction(trsresbytes, 0)
	if err != nil {
		result["err"] = err.Error()
		return result
	}

	// 解析 actions
	var allactions = trsres.GetActions()
	var actions_ary []string
	var actions_strings = ""
	for _, act := range allactions {
		var kind = act.Kind()
		actstr := fmt.Sprintf(`{"k":%d`, kind)
		if kind == 1 {
			acc := act.(*actions.Action_1_SimpleTransfer)
			actstr += fmt.Sprintf(`,"to":"%s","amount":"%s"`,
				acc.Address.ToReadable(),
				acc.Amount.ToFinString(),
			)
		} else if kind == 2 {
			acc := act.(*actions.Action_2_OpenPaymentChannel)
			actstr += fmt.Sprintf(`,"channel_id":"%s","left_addr":"%s","left_amt":"%s","right_addr":"%s","right_amt":"%s"`,
				hex.EncodeToString(acc.ChannelId),
				acc.LeftAddress.ToReadable(),
				acc.LeftAmount.ToFinString(),
				acc.RightAddress.ToReadable(),
				acc.RightAmount.ToFinString(),
			)
		} else if kind == 3 {
			acc := act.(*actions.Action_3_ClosePaymentChannel)
			actstr += fmt.Sprintf(`,"channel_id":"%s"`,
				hex.EncodeToString(acc.ChannelId),
			)
		} else if kind == 4 {
			acc := act.(*actions.Action_4_DiamondCreate)
			actstr += fmt.Sprintf(`,"number":"%s","name":"%s","address":"%s"`,
				acc.Address.ToReadable(),
				acc.Diamond,
				acc.Address.ToReadable(),
			)
		}
		actstr += "}"
		actions_ary = append(actions_ary, actstr)
	}
	actions_strings = strings.Join(actions_ary, ",")
	// 交易返回数据
	txaddr := fields.Address(trsres.GetAddress())
	var txfee = trsres.GetFee()
	result["jsondata"] = fmt.Sprintf(
		`{"block":{"height":%d,"timestamp":%d},"type":%d,"address":"%s","fee":"%s","timestamp":%d,"actioncount":%d,"actions":[%s]`,
		blkhei,
		trsres.GetTimestamp(),
		trsres.Type(),
		txaddr.ToReadable(), // 主地址
		txfee.ToFinString(),
		trsres.GetTimestamp(),
		len(allactions),
		actions_strings,
	)

	// 收尾并返回
	result["jsondata"] += "}"
	return result
}
