package main

import (
	"fmt"
	// "io"
	// "os"
	// "os/exec"
	// "strings"
	// "time"
	"strconv"
	"encoding/hex"
	"regexp"
	"hash/fnv"
	"github.com/ethereum/collector"
	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary
var bytecodeHash_map map[string]map[string]int    // store bytecodeHash and jump list


// get func jump table
func GetJumpTable(runtime_bytecode string) map[string]int {
	len_bytecode := len(runtime_bytecode)
	var opcode_value_list [][2] string   // store opcode and value
	x := 0// record the current place
	reg, err  := regexp.Compile("^[6-7]?[0-9a-fA-F]$")
	if err != nil{
		fmt.Println(err.Error())
	}

	// get opcode and value list
	for x < len_bytecode {
		if (x + 2) <= len_bytecode{
			current_opcode := runtime_bytecode[x : x + 2]
			if reg.MatchString(current_opcode){
				current_bytecode_10, err1 := strconv.ParseInt(current_opcode, 16, 0)
				if err1 != nil{
					fmt.Println(err1.Error())
				}
				len_value := int((current_bytecode_10 - 95) * 2)
				current_value := ""
				if (x + 2 + len_value) <= len_bytecode{
					current_value = runtime_bytecode[x + 2 : x + 2 + len_value]
				}else{
					current_value = runtime_bytecode[x + 2 :]
				}
				var temp_opcode_value_list = [2] string{current_opcode, current_value}
				opcode_value_list = append(opcode_value_list, temp_opcode_value_list)
				x =  x + 2 + len_value
			}else{
				var temp_opcode_value_list = [2] string{current_opcode, "0"}
				opcode_value_list = append(opcode_value_list, temp_opcode_value_list)
				x =  x + 2
			}
		}else{
			break
		}
	}

	// fmt.Println(opcode_value_list)

	// get jump table
	jump_table_dict := make(map[string]int)
	len_opcode_value := len(opcode_value_list)
	for i := 0; i < len_opcode_value; i++ {
		if opcode_value_list[i][0] == "80"{  // dup1
			if ((i + 1) < len_opcode_value) && ((opcode_value_list[i + 1][0] == "62") || (opcode_value_list[i + 1][0] == "63")){  // push3,push4
				if ((i + 2) < len_opcode_value) && (opcode_value_list[i + 2][0] == "14"){ //EQ
					if ((i + 3) < len_opcode_value) && (reg.MatchString(opcode_value_list[i + 3][0])){ // push n
						if ((i + 4) < len_opcode_value) && ((opcode_value_list[i + 4][0] == "56") || (opcode_value_list[i + 4][0] == "57")){  // jumpi,jump
							//methodid := opcode_value_list[i + 1][1].zfill(8)
							methodid := fmt.Sprintf("%08s", opcode_value_list[i + 1][1])
							jump_table_dict[methodid] = 0
						}
					}
				}
			}
		}else{
			if (opcode_value_list[i][0] == "62") || (opcode_value_list[i][0] == "63"){ // push3,push4
				if ((i + 1) < len_opcode_value) && (opcode_value_list[i + 1][0] == "81"){  // dup2
					if ((i + 2) < len_opcode_value) && (opcode_value_list[i + 2][0] == "14"){  // EQ
						if ((i + 3) < len_opcode_value) && (reg.MatchString(opcode_value_list[i + 3][0])){  // push n
							if ((i + 4) < len_opcode_value) && ((opcode_value_list[i + 4][0] == "56") || (opcode_value_list[i + 4][0] == "57")){  // jumpi,jump
								//methodid := opcode_value_list[i][1].zfill(8)
								methodid := fmt.Sprintf("%08s", opcode_value_list[i][1])
								jump_table_dict[methodid] = 0
							}
						}else{  // 处理一下assert的情况
							if ((i + 3) < len_opcode_value) && (opcode_value_list[i + 3][0] == "15"){  // iszero
								if ((i + 4) < len_opcode_value) && (reg.MatchString(opcode_value_list[i + 4][0])){  // push n
									if ((i + 5) < len_opcode_value) && ((opcode_value_list[i + 5][0] == "56") || (opcode_value_list[i + 5][0] == "57")){  // jumpi,jump
										//methodid := opcode_value_list[i][1].zfill(8)
										methodid := fmt.Sprintf("%08s", opcode_value_list[i][1])
										jump_table_dict[methodid] = 0
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return jump_table_dict
}



// make hash for bytecode
func Fnvhash(bytecode []byte) string{
	result := fnv.New64a()
	result.Write(bytecode)
	return hex.EncodeToString(result.Sum(nil))
}

type RegisterInfo struct {
	PluginName string   `json:"pluginname"`
	OpCode     map[string]string `json:"option"`
}

// 插件入口函数
func Register() []byte {
	// fmt.Println("plugin run")
	var data = RegisterInfo{
		PluginName: "P2",
		OpCode: map[string]string{"IAL_BYTECODE":"Handle_BYTECODE","IAL_INVOKE":"Handle_INVOKE"},
	}
	bytecodeHash_map = make(map[string]map[string]int)

	retInfo, err := json.Marshal(&data)
	if err != nil {
		fmt.Println(err)
		// panic(err)
	}
	//fmt.Println(b)

	return retInfo
}

// judge if the method is in the jump dict
func InJump(input string, bytecodeHash string) string {
	ll := len(input)
	var methodid = ""
	if ll >= 8{
		temp_ll := ll - 8
		if temp_ll % 64 == 0{
			methodid = input[0:8]
		}else{
			return "0"
		}
	}else{
		return "0"
	}

	_, ok := bytecodeHash_map[bytecodeHash]
	if ok{
		jump_table_dict := bytecodeHash_map[bytecodeHash]
		if _, ok := jump_table_dict[methodid]; !ok{
			return "1"
		}
	}else{
		return "0"
	}
	return "0"
}

// add bytecodeHash to dict
func add_to_dict(runtimecode []byte) {
	var runtimecode_str string = ""
	runtimecode_str = hex.EncodeToString(runtimecode)
	if (len(runtimecode_str) > 0){
		bytecodeHash := Fnvhash(runtimecode)
		if _,ok := bytecodeHash_map[bytecodeHash]; !ok{
			jump_table_dict := GetJumpTable(runtimecode_str)
			bytecodeHash_map[bytecodeHash] = jump_table_dict
		}
		
	}
}

func Handle_INVOKE(m *collector.AllCollector) (byte ,string){
	if m.TransInfo.CallType == "CALL"{   // external call, get contract name and input, check if the method is in the jumptable
		input := hex.EncodeToString(m.TransInfo.CallInfo.InputData)
		if len(m.TransInfo.CallInfo.ContractCode) > 0{
			bytecodeHash := Fnvhash(m.TransInfo.CallInfo.ContractCode)
			get_result := InJump(input, bytecodeHash)
			if get_result == "1"{
				return 0x01,input[0:8]
			}	
		}					
	}
	return 0x00,""
}

func Handle_BYTECODE(m *collector.AllCollector) (byte ,string) {
	if m.TransInfo.CallType == "CREATE"{
		runtimecode := m.TransInfo.CreateInfo.ContractRuntimeCode
		if len(runtimecode) > 0{
			add_to_dict(runtimecode)
		}
	}
	return 0x00,""
}

