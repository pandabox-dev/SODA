
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

type RunOpCode struct {
	MethodName string   `json:"methodname"`
	OpCode     []string `json:"option"`
}

func Run() []byte {
	var data = RunOpCode{
		MethodName: "detect_balance",
		OpCode: []string{"TXSTART", "NOT", "LT", "GT", "SLT", "SGT" ,"BALANCE", "EQ", "ISZERO"},
	}
	initial()

	b, err := json.Marshal(&data)
	if err != nil {
		return nil
	}

	return b
}


func initial() {
	cmp_flag = 0
	balance = ""
	balance_layer = 0
}

func Recv(m *collector.CollectorDataT) (byte ,string) {
	if m.Option == "TXSTART" {
		initial()
		return 0x00,""
	}

	if m.Option == "BALANCE" {
		if m.InsInfo.OpResult != ""{
			cmp_flag = 0
			balance_layer = m.InsInfo.CallLayer
			balance = m.InsInfo.OpResult
		}
		return 0x00,""
	}

	if m.Option == "EQ"{
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

	if m.Option == "NOT" || m.Option == "LT" || m.Option == "GT" || m.Option == "SLT" || m.Option == "SGT" || m.Option == "ISZERO"{
		cmp_flag = 1
		return 0x00,""
	}

	return 0x00,""
}
