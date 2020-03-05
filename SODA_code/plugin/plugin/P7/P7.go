
package main

import (
	// "../../pluginlog"
	// "encoding/hex"
	"github.com/ethereum/collector"
	// "math/big"
	// "fmt"
	"github.com/json-iterator/go"
	// "strings"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var (
	balance 	    string   // store balance result
	cmp_flag 		int
	balance_layer	int
)

type RegisterInfo struct {
	PluginName string   `json:"pluginname"`
	OpCode     map[string]string `json:"option"`
}

func Register() []byte {
	var data = RegisterInfo{
		PluginName: "P7",
		OpCode: map[string]string{"TXSTART":"Handle_TXSTART", "NOT":"Handle_COMPARE", "LT":"Handle_COMPARE", "GT":"Handle_COMPARE", "SLT":"Handle_COMPARE", "SGT":"Handle_COMPARE", "BALANCE":"Handle_BALANCE", "EQ":"Handle_EQ", "ISZERO":"Handle_COMPARE"},
	}
	initial()

	retInfo, err := json.Marshal(&data)
	if err != nil {
		return nil
	}

	return retInfo
}


func initial() {
	cmp_flag = 0
	balance = ""
	balance_layer = 0
}

func Handle_TXSTART(m *collector.AllCollector) (byte ,string){
	initial()
	return 0x00,""
}

func Handle_BALANCE(m *collector.AllCollector) (byte ,string){
	if m.InsInfo.OpInOut.OpResult != ""{  // add tutu
		cmp_flag = 0
		balance_layer = m.InsInfo.CallLayer
		balance = m.InsInfo.OpInOut.OpResult  //add tutu
	}
	return 0x00,""
}

func Handle_EQ(m *collector.AllCollector) (byte ,string){
	if len(m.InsInfo.OpInOut.OpArgs) == 2 && cmp_flag == 0{  //add tutu
		result_flag := 0
		current_layer := m.InsInfo.CallLayer
		if current_layer == balance_layer{
			for _,eq_str := range m.InsInfo.OpInOut.OpArgs {  //add tutu
				if balance == eq_str{
					result_flag = 1
				}
			}
		}
		if result_flag == 1 {
			return 0x01,""
		}
	}
	balance = ""
	return 0x00,""
}

func Handle_COMPARE(m *collector.AllCollector) (byte ,string) {
	cmp_flag = 1
	return 0x00,""
}
