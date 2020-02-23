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

type RegisterInfo struct {
	PluginName string   `json:"pluginname"`
	OpCode     map[string]string `json:"option"`
}

func Register() []byte {
	// fmt.Println("enter run")
	var data = RegisterInfo{
		PluginName: "P8",
		OpCode: map[string]string{"TXSTART":"handle_TXSTART", "IAL_COMPARISON":"handle_COMPARISON", "NUMBER":"handle_NUMBERTIME", "TIMESTAMP":"handle_NUMBERTIME"},
	}

	dependecy_map = make(map[int]map[string]int)

	retInfo, err := json.Marshal(&data)
	if err != nil {
		return nil
	}

	return retInfo
}

func handle_TXSTART(m *collector.CollectorDataT) (byte ,string){
	dependecy_map = make(map[int]map[string]int)
	return 0x00,""
}

func handle_NUMBERTIME(m *collector.CollectorDataT) (byte ,string){
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

func handle_COMPARISON(m *collector.CollectorDataT) (byte ,string){
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
