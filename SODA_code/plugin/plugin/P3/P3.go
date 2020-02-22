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


type RunOpCode struct {
	MethodName string   `json:"methodname"`
	OpCode     []string `json:"option"`
}

// 插件入口函数
func Run() []byte {
	// fmt.Println("plugin run")
	var data = RunOpCode{
		MethodName: "P3",
		OpCode: []string{"handleIAL_INVOKE"},
	}

	b, err := json.Marshal(&data)
	if err != nil {
		fmt.Println(err)
		// panic(err)
	}
	//fmt.Println(b)

	return b
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
func Recv(m *collector.CollectorDataT) (byte ,string) {
	
	if m.Option == "handleIAL_INVOKE" {
		if m.TransInfo.CallType == "CALL"{
			// external call, get contract name and input, check if the method is in the jumptable
			input := hex.EncodeToString(m.TransInfo.CallInfo.InputData)
			result := check_length(input)
			if result == "1"{
				return 0x01,input
			}
		}
	}

	return 0x00,""
}
