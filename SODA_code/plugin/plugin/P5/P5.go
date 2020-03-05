package main

import (
	// "os"
	"fmt"
	"strings"
	// "bufio"
	// "strconv"
	"hash/fnv"
	"encoding/hex"
	"github.com/json-iterator/go"
	"github.com/ethereum/collector"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary
var bytecodeHash_map map[string]map[uint64]int    // store contract and pc list
var contract_map map[string]string

type RegisterInfo struct {
	PluginName string   `json:"pluginname"`
	OpCode     map[string]string `json:"option"`
}

func Register() []byte {
	
	var data = RegisterInfo{
		PluginName: "P5",
		OpCode: map[string]string{"IAL_BYTECODE":"Handle_BYTECODE","TRANS_CALL":"Handle_CALLINFO","TRANS_CALLCODE":"Handle_CALLINFO","TRANS_DELEGATECALL":"Handle_CALLINFO"},
	}

	contract_map = make(map[string]string)
	bytecodeHash_map = make(map[string]map[uint64]int)
	// fmt.Println(bytecodeHash_map)
	retInfo, err := json.Marshal(&data)
	if err != nil {
		return []byte{}
	}

	return retInfo
}

func check_return_value(runtimecode []byte) map[uint64]int {
	length := len(runtimecode)
	result := make(map[uint64]int)
	for i := 0; i < length; i++ {
		//96-127:push1 - push32
		if v := int(runtimecode[i]); v >= 96 && v <= 127 {
			i += v - 95
		//241:call 242:callcode
		} else if runtimecode[i] == 241 || runtimecode[i] == 242 || runtimecode[i] == 244 {
			
			if i + 1 >= length{
				continue
			}

			if runtimecode[i+1]  == 21{
				continue
			}

			swap_num := 0
			//144-159:swap1 - swap16

			if s := int(runtimecode[i+1]) ;s >= 144 && s <= 159 {
				swap_num = s-143
			}else{

				check_flag := 0

				for j := i+1;j < length;j++{
					if j - i > 10{
						break
					}
					opcode1 := byte(0)
					opcode2 := byte(0)
					opcode3 := byte(0)

					if j <length{
						opcode1 = runtimecode[j]
					}
					if j+1 <length{
						opcode2 = runtimecode[j+1]
					}
					if j+2 <length {
						opcode3 = runtimecode[j+2]
					}

					if opcode1 == 128 && opcode2 == 21{
						check_flag = 1
						break
					}
					
					if opcode1 == 96 && opcode2 == 129 && opcode3 == 20{
						check_flag = 1
						break
					}
 
				}

				if check_flag == 0{
					result[uint64(i)] = 0
					continue
				}
				
			}

			pop_flag := 0
			for j:=0;j < swap_num;j++{
				if i + j + 2 >= length{
					pop_flag = 1
					break
				}
				if runtimecode[i+j+2] != 80{ //80:pop
					pop_flag = 1
				}
			}
			if pop_flag == 1{
				result[uint64(i)] = 0
				continue
			}

			opcode1 := byte(0)
			opcode2 := byte(0)
			opcode3 := byte(0)
			opcode4 := byte(0)
			opcode5 := byte(0)
			if i+swap_num+2 <length{
				opcode1 = runtimecode[i+swap_num+2]
			}
			if i+swap_num+3 <length{
				opcode2 = runtimecode[i+swap_num+3]
			}
			if i+swap_num+4 <length{
				opcode3 = runtimecode[i+swap_num+4]
			}
			if i+swap_num+5 <length{
				opcode4 = runtimecode[i+swap_num+5]
			}
			if i+swap_num+6 <length{
				opcode4 = runtimecode[i+swap_num+6]
			}

			check_flag := 0
			//21:ISZERO 128-143:DUP1-DUP16 91:JUMPDEST 96:PUSH1 20:EQ
			if opcode1 == 21{
				check_flag = 1
			}
			if opcode1>=128&&opcode1<=143{
				if opcode2 == 21{
					check_flag =1
				}
			}
			if opcode1 == 91{
				if opcode2 == 21{
					check_flag =1
				}
				if opcode2>=128&&opcode2<=143{
					if opcode3 == 21{
						check_flag = 1
					}
				}
			}
			if opcode1 == 96{
				if opcode3 == 20{
					check_flag = 1
				}
			}

			if opcode1 == 144 && opcode2 == 80 && opcode3 == 128 && opcode4 == 21 {
				check_flag = 1		
			}

			if opcode1 == 61 && opcode2 == 128 && opcode3 == 96 && opcode4 == 129 && opcode5 == 20{
				check_flag = 1		
			}

			if check_flag == 0{
				result[uint64(i)] = 0
				continue
			}

		}
	}
	return result
}

func Fnvhash(bytecode []byte) string{
	result := fnv.New64a()
	result.Write(bytecode)
	return hex.EncodeToString(result.Sum(nil))
}

func PcInDict(pc uint64, bytecodeHash string) int {
	if _, ok := bytecodeHash_map[bytecodeHash];ok{
		pc_dict := bytecodeHash_map[bytecodeHash]
		if _, eok := pc_dict[pc]; eok{
			return 1
		}
	}
	return 0
}

func Handle_BYTECODE(m *collector.AllCollector) (byte ,string){
	if m.TransInfo.CallType == "CREATE" {
		contract := m.TransInfo.To
		contract = strings.ToLower(contract)
		runtimecode := m.TransInfo.CreateInfo.ContractRuntimeCode
		if len(runtimecode) > 0{
			bytecodeHash := Fnvhash(runtimecode)
			if _, ok := bytecodeHash_map[bytecodeHash];!ok{
				pc_dict := check_return_value(runtimecode)
				bytecodeHash_map[bytecodeHash] = pc_dict
				contract_map[contract] = bytecodeHash
			}else{
				contract_map[contract] = bytecodeHash
			}
		}
	}	
	return 0X00,""
}


func Handle_CALLINFO(m *collector.AllCollector) (byte ,string){
	if m.TransInfo.CallType == "CALL" {
		contract := m.TransInfo.From 
		toaddr := m.TransInfo.To
		contract = strings.ToLower(contract)
		pc := m.TransInfo.Pc 
		layer := m.TransInfo.CallLayer
		bytecodeHash := contract_map[contract]
		get_result := PcInDict(pc, bytecodeHash)
		if get_result == 1 && m.TransInfo.IsSuccess==false && len(m.TransInfo.CallInfo.ContractCode)>0{
			return 0x01 , contract + "#" + toaddr +"#"+fmt.Sprintf("%v", layer)+"#"+fmt.Sprintf("%v", pc)
		}
	}
	return 0x00,""
}





