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


var	dependecy_map map[int]map[string]int   // store timestamp and number result

type RunOpCode struct {
	MethodName string   `json:"methodname"`
	OpCode     []string `json:"option"`
}

func Run() []byte {
	// fmt.Println("enter run")
	var data = RunOpCode{
		MethodName: "detect_dependency",
		OpCode: []string{"TXSTART", "handleIAL_COMPARISON", "NUMBER", "TIMESTAMP"},
	}

	dependecy_map = make(map[int]map[string]int)

	b, err := json.Marshal(&data)
	if err != nil {
		return nil
	}

	return b
}

func Recv(m *collector.CollectorDataT) (byte ,string) {
	// 记录下当前的txhash，块高以及时间戳
	if m.Option == "TXSTART" {
		dependecy_map = make(map[int]map[string]int)
		return 0x00,""
	}

	if m.Option == "NUMBER" || m.Option == "TIMESTAMP"{
		current_layer := m.InsInfo.CallLayer
		need_result := m.InsInfo.OpResult
		if _,ok := dependecy_map[current_layer]; ok{
			dependecy_map[current_layer][need_result] = 0
		}else{
			temp_map := make(map[string]int)
			temp_map[need_result] = 0
			dependecy_map[current_layer] = temp_map
		}
		return 0x00,""
	}

	if m.Option == "handleIAL_COMPARISON" {
		current_layer := m.InsInfo.CallLayer
		if _, ok := dependecy_map[current_layer]; ok{
			for _,i_str := range m.InsInfo.OpArgs{
				if _, ok1 := dependecy_map[current_layer][i_str]; ok1 {
					return 0x01,""
				}
			}
		}
		
		return 0x00,""
	}

	return 0x00,""
}
