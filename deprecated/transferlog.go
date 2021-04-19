package rpc

import (
	"github.com/hacash/core/actions"
	"github.com/hacash/core/fields"
	rpc "github.com/hacash/service/server"
	"strconv"
	"strings"
)

// 扫描区块 获取所有转账信息
func (api *DeprecatedApiService) getAllTransferLogByBlockHeight(params map[string]string) map[string]string {
	result := make(map[string]string)
	block_height_str, ok1 := params["block_height"]
	if !ok1 {
		result["err"] = "param block_height must."
		return result
	}
	block_height, err2 := strconv.ParseUint(block_height_str, 10, 0)
	if err2 != nil {
		result["err"] = "param block_height format error."
		return result
	}

	state := api.blockchain.State()

	lastest, e3 := state.ReadLastestBlockHeadAndMeta()
	if e3 != nil {
		result["err"] = e3.Error()
		return result
	}

	// 判断区块高度
	if block_height <= 0 || block_height > lastest.GetHeight() {
		result["err"] = "block height not find."
		result["ret"] = "1" // 返回错误码
		return result
	}

	// 查询区块
	tarblock, e := rpc.LoadBlockWithCache(api.backend.BlockChain().State(), block_height)
	if e != nil {
		result["err"] = "read block data error."
		return result
	}

	// 开始扫描区块
	allTransferLogs := make([]string, 0, 4)
	transactions := tarblock.GetTransactions()
	for _, v := range transactions {
		if 0 == v.Type() { // coinbase
			continue
		}
		from := fields.Address(v.GetAddress())
		for _, act := range v.GetActions() {
			if 1 == act.Kind() { // 类型为普通转账
				act_k1 := act.(*actions.Action_1_SimpleTransfer)
				allTransferLogs = append(allTransferLogs,
					from.ToReadable()+"|"+
						act_k1.ToAddress.ToReadable()+"|"+
						act_k1.Amount.ToFinString())
			}
		}
	}

	datasstr := strings.Join(allTransferLogs, "\",\"")
	if len(datasstr) > 0 {
		datasstr = "\"" + datasstr + "\""
	}

	// 返回
	result["jsondata"] = `{"timestamp":` + strconv.FormatUint(tarblock.GetTimestamp(), 10) + `,"datas":[` + datasstr + `]}`
	return result
}
