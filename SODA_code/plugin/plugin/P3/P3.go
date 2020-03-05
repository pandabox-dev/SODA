package main

import (
	"fmt"
	// "io"
	// "os"
	// "os/exec"
	"strings"
	// "time"
	// "strconv"
	"encoding/hex"
	// "regexp"
	// "hash/fnv"
	"github.com/ethereum/collector"
	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary


type RegisterInfo struct {
	PluginName string   `json:"pluginname"`
	OpCode     map[string]string `json:"option"`
}

// 插件入口函数
func Register() []byte {
	// fmt.Println("plugin run")
	var data = RegisterInfo{
		PluginName: "P3",
		OpCode: map[string]string{"IAL_INVOKE":"Handle_INVOKE"},
	}

	retInfo, err := json.Marshal(&data)
	if err != nil {
		fmt.Println(err)
		// panic(err)
	}
	//fmt.Println(b)

	return retInfo
}

// judge the lenth of the input
func check_length(input string) string {
	ll := len(input)
	// var methodid = ""
	if ll >= 8{
		methodid := strings.ToLower(input[0:8])
		temp_ll := ll - 8
		if methodid == "a9059cbb"{
			if temp_ll < 128{
				return "1"
			}
		}
		if methodid == "23b872dd"{
			if temp_ll < 192{
				return "1"
			}
		}
	}
	return "0"
}

//return 0X01 结束
func Handle_INVOKE(m *collector.AllCollector) (byte ,string) {
	if m.TransInfo.CallType == "CALL"{
		// external call, get contract name and input, check if the method is in the jumptable
		input := hex.EncodeToString(m.TransInfo.CallInfo.InputData)
		result := check_length(input)
		if result == "1"{
			return 0x01,input
		}
	}

	return 0x00,""
}
