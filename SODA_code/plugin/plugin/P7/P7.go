
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
		OpCode: map[string]string{"TXSTART":"handle_TXSTART", "NOT":"handle_COMPARE", "LT":"handle_COMPARE", "GT":"handle_COMPARE", "SLT":"handle_COMPARE", "SGT":"handle_COMPARE", "BALANCE":"handle_BALANCE", "EQ":"handle_EQ", "ISZERO":"handle_COMPARE"},
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

func handle_TXSTART(m *collector.CollectorDataT) (byte ,string){
	initial()
	return 0x00,""
}

func handle_BALANCE(m *collector.CollectorDataT) (byte ,string){
	if m.InsInfo.OpResult != ""{
		cmp_flag = 0
		balance_layer = m.InsInfo.CallLayer
		balance = m.InsInfo.OpResult
	}
	return 0x00,""
}

func handle_EQ(m *collector.CollectorDataT) (byte ,string){
	if len(m.InsInfo.OpArgs) == 2 && cmp_flag == 0{ 
		result_flag := 0
		current_layer := m.InsInfo.CallLayer
		if current_layer == balance_layer{
			for _,eq_str := range m.InsInfo.OpArgs {
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

func handle_COMPARE(m *collector.CollectorDataT) (byte ,string) {
	cmp_flag = 1
	return 0x00,""
}
