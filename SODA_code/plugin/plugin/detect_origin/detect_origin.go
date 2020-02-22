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

type RunOpCode struct {
	MethodName string   `json:"methodname"`
	OpCode     []string `json:"option"`
}

func Run() []byte {
	// fmt.Println("enter run")
	var data = RunOpCode{
		MethodName: "detect_origin",
		OpCode: []string{"TXSTART", "EQ", "ORIGIN","CALLSTART","CALLCODESTART","DELEGATECALLSTART","STATICCALLSTART"},
	}
	origin_map = make(map[int]map[string]int)
	sender_map = make(map[int]string)
	// logger.InitialFileLog("./tx_err_log/detect_origin_new/detect_origin")
	b, err := json.Marshal(&data)
	if err != nil {
		return nil
	}

	return b
}

func Recv(m *collector.CollectorDataT) (byte ,string){
	// clear current origin_map
	if m.Option == "TXSTART" {
		origin_map = make(map[int]map[string]int)
		sender_map = make(map[int]string)
		current_layer := m.TransInfo.CallLayer
		sender := m.TransInfo.From
		sender_map[current_layer] = sender
		// txhash = m.ExternalInfo.TxHash
		return 0x00,""
	}

	if m.Option == "CALLSTART" || m.Option == "CALLCODESTART" || m.Option =="DELEGATECALLSTART" || m.Option == "STATICCALLSTART"{
		current_layer := m.InsInfo.CallLayer
		sender := m.InsInfo.From
		sender_map[current_layer] = sender
		return 0x00,""
	}

	if m.Option == "ORIGIN" {
		current_layer := m.InsInfo.CallLayer
		origin_addr := m.InsInfo.OpResult
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

	if m.Option == "EQ" {
		current_layer := m.InsInfo.CallLayer
		if _, ok := origin_map[current_layer]; ok{  // eq appear where origin appear
			for _,i_str := range m.InsInfo.OpArgs{
				if _, ok1 := origin_map[current_layer][i_str]; ok1{
					origin_addr_big,_ := new(big.Int).SetString(i_str,10)
					origin_addr_16 := "0x" + fmt.Sprintf("%040x",origin_addr_big)
					origin_addr_16 = strings.ToLower(origin_addr_16)
					// write_str := origin_addr_16
					if m.InsInfo.OpResult == "1"{
						current_sender := strings.ToLower(sender_map[current_layer])
						if origin_addr_16 != current_sender{
							write_str := current_sender + "#" + origin_addr_16
							return 0x01,write_str
						}
					}
					// return 0x01,write_str
				}
			}
		}
		return 0x00,""
	}

	return 0x00,""
}
