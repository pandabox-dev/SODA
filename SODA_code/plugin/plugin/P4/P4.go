package main

import (
	// "../../pluginlog"
	// "encoding/hex"
	"github.com/ethereum/collector"
	"math/big"
	"fmt"
	"github.com/json-iterator/go"
	"strings"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary
// var logger pluginlog.ErrTxLog

var	origin_map map[int]map[string]int   // store origin result
// var txhash string
var sender_map map[int]string  //store sender of each layer

type RegisterInfo struct {
	PluginName string   `json:"pluginname"`
	OpCode     map[string]string `json:"option"`
}

func Register() []byte {
	// fmt.Println("enter run")
	var data = RegisterInfo{
		PluginName: "P4",
		OpCode: map[string]string{"EXTERNALINFOSTART":"Handle_EXTERNALINFOSTART", "EQ":"Handle_EQ", "ORIGIN":"Handle_ORIGIN","CALLSTART":"Handle_CALLINFO","CALLCODESTART":"Handle_CALLINFO","DELEGATECALLSTART":"Handle_CALLINFO","STATICCALLSTART":"Handle_CALLINFO"},
	}
	origin_map = make(map[int]map[string]int)
	sender_map = make(map[int]string)
	// logger.InitialFileLog("./tx_err_log/detect_origin_new/detect_origin")
	retInfo, err := json.Marshal(&data)
	if err != nil {
		return nil
	}

	return retInfo
}


func Handle_EXTERNALINFOSTART(m *collector.AllCollector) (byte ,string){
	origin_map = make(map[int]map[string]int)
	sender_map = make(map[int]string)
	current_layer := m.TransInfo.CallLayer
	sender := m.TransInfo.From
	sender_map[current_layer] = sender
	// txhash = m.ExternalInfo.TxHash
	return 0x00,""
}

func Handle_CALLINFO(m *collector.AllCollector) (byte ,string){
	current_layer := m.InsInfo.CallLayer
	sender := m.InsInfo.AccountValue.FromAddr  //add tutu
	sender_map[current_layer] = sender
	return 0x00,""
}

func Handle_EQ(m *collector.AllCollector) (byte ,string){
	current_layer := m.InsInfo.CallLayer
	if _, ok := origin_map[current_layer]; ok{  // eq appear where origin appear
		for _,i_str := range m.InsInfo.OpInOut.OpArgs{  //add tutu
			if _, ok1 := origin_map[current_layer][i_str]; ok1{
				origin_addr_big,_ := new(big.Int).SetString(i_str,10)
				origin_addr_16 := "0x" + fmt.Sprintf("%040x",origin_addr_big)
				origin_addr_16 = strings.ToLower(origin_addr_16)
				write_str := origin_addr_16
				if m.InsInfo.OpInOut.OpResult == "1"{  //add tutu
					current_sender := strings.ToLower(sender_map[current_layer])
					if origin_addr_16 != current_sender{
						write_str = current_sender + "#" + origin_addr_16
						return 0x01,write_str
					}
				}
			}
		}
	}
	return 0x00,""
}

func Handle_ORIGIN(m *collector.AllCollector) (byte ,string){
	current_layer := m.InsInfo.CallLayer
	origin_addr := m.InsInfo.OpInOut.OpResult   //add tutu
	// origin_addr_big,_ := new(big.Int).SetString(origin_addr,10)
	// origin_addr_16 := "0x" + fmt.Sprintf("%040x",origin_addr_big)
	// origin_addr_16 = strings.ToLower(origin_addr_16)
	if _,ok := origin_map[current_layer]; ok{
		origin_map[current_layer][origin_addr] = 0
	}else{
		temp_map := make(map[string]int)
		temp_map[origin_addr] = 0
		origin_map[current_layer] = temp_map
	}
	return 0x00,""
}
