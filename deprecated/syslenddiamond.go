package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/stores"
	"strconv"
)

//////////////////////////////////////////////////////////////

func (api *DeprecatedApiService) getSystemLendingDiamond(params map[string]string) map[string]string {
	result := make(map[string]string)
	idstr, ok1 := params["id"]
	if !ok1 {
		result["err"] = "params id must."
		return result
	}

	dataid, e0 := hex.DecodeString(idstr)
	if e0 != nil {
		result["fail"] = "id format error."
		return result
	}

	if len(dataid) != stores.DiamondSyslendIdLength {
		result["fail"] = "id length error."
		return result
	}

	state := api.blockchain.GetChainEngineKernel().StateRead()
	stoobj, _ := state.DiamondSystemLending(fields.DiamondSyslendId(dataid))
	if stoobj == nil {
		result["fail"] = "not find."
		return result
	}

	// Return details
	result["is_ransomed"] = strconv.Itoa(int(stoobj.IsRansomed.Value()))
	result["create_block_height"] = strconv.FormatUint(uint64(stoobj.CreateBlockHeight), 10)
	result["main_address"] = stoobj.MainAddress.ToReadable()
	result["borrow_period"] = strconv.Itoa(int(stoobj.BorrowPeriod))
	result["interest_rate"] = fmt.Sprintf("%.1f%%", float64(stoobj.BorrowPeriod)*0.5) // Unit percentage
	result["mortgage_time"] = fmt.Sprintf("%.1f days", float64(stoobj.BorrowPeriod)*10000/288)
	result["diamonds"] = stoobj.MortgageDiamondList.SerializeHACDlistToCommaSplitString() // Diamond list
	result["loan_amount"] = fmt.Sprintf("ㄜ%d:248", stoobj.LoanTotalAmountMei)

	// Show return status
	if stoobj.IsRansomed.Check() {
		result["ransom_block_height"] = strconv.FormatUint(uint64(stoobj.RansomBlockHeight), 10)
		result["ransom_amount"] = stoobj.RansomAmount.ToFinString()
		result["ransom_address_if_public_operation"] = stoobj.RansomAddressIfPublicOperation.ShowReadableOrEmpty() // Show address if present
	}

	// return
	return result
}
