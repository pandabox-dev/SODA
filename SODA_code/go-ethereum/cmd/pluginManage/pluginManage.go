package pluginManage

//add new file

import (
	"github.com/ethereum/collector"
	"github.com/ethereum/go-ethereum/cmd/pluginManage/Monitor"
	"regexp"
	"strconv"
	"strings"
	// "fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/tingrong"
	"github.com/ethereum/go-ethereum/fei"
)

//2019.03.01 version plugin

type monitorType struct{
	mtor 	*monitor.MonitorMethod
	option 	string
}

type PluginManages struct {
	plugins map[string][]monitorType
	
}

var clearvalue []monitorType

func NewPluginManages() *PluginManages {
	return &PluginManages{
		make(map[string][]monitorType),
	}
}

func (plg *PluginManages) HasFuncRegistered() bool {
	return len(plg.plugins) > 0
}

func (plg *PluginManages) regexpString(value string, mtor *monitor.MonitorMethod) {
	strValue := plg.findStrFrom(value, `[a-z_A-z]+`)
	//PUSH8-20*
	if len(strValue) == 1 {
		srrstr := plg.findStrFrom(value, `\d+\-\d+`)
		if len(srrstr) >= 1 {
			arr := strings.Split(srrstr[0], "-")
			if len(arr) == 2 {
				max, _ := strconv.Atoi(arr[1])
				min, _ := strconv.Atoi(arr[0])
				if min > max {
					temp := max
					max = min
					min = temp
				}

				for ; min <= max; min++ {
					key := strValue[0] + strconv.Itoa(min)
					if IsExist(key) {
						monitorT := monitorType{} 
						monitorT.mtor  = mtor
						monitorT.option  = key
						plg.plugins[key] = append(plg.plugins[key], monitorT)
					}
				}
			}
		}

		//PUSH*
		spstr := plg.findStrFrom(value, `\*`)
		if len(spstr) == 1 && spstr[0] == "*" {
			for index := 0; index < 30; index++ {
				key := strValue[0] + strconv.Itoa(index)
				if IsExist(key) {
					monitorT := monitorType{} 
					monitorT.mtor  = mtor
					monitorT.option  = key
					plg.plugins[key] = append(plg.plugins[key], monitorT)
				}
			}
		}
	}
}

func (plg *PluginManages) findStrFrom(str string, expr string) []string {
	special, _ := regexp.Compile(expr)
	return special.FindAllString(str, -1)
}

func (plg *PluginManages) RegisterOption(opcodes []string, mtor *monitor.MonitorMethod) {
	mtor.SetStatus(false)
	for _,value := range opcodes {
		if IsExist(value) {
			switch value{
			case "handleIAL_BYTECODE":
				monitorT := monitorType{} 
				monitorT.mtor = mtor
				monitorT.option ="handleIAL_BYTECODE"
				plg.plugins["EXTERNALINFOSTART"] = append(plg.plugins["EXTERNALINFOSTART"], monitorT)
				plg.plugins["EXTERNALINFOEND"] = append(plg.plugins["EXTERNALINFOEND"], monitorT)
				plg.plugins["TRANS_CREATE"] = append(plg.plugins["TRANS_CREATE"], monitorT)
				plg.plugins["TRANS_CREATE2"] = append(plg.plugins["TRANS_CREATE2"], monitorT)
			case "handleIAL_INVOKE":
				monitorT := monitorType{} 
				monitorT.mtor  = mtor
				monitorT.option  ="handleIAL_INVOKE"
				plg.plugins["EXTERNALINFOSTART"] = append(plg.plugins["EXTERNALINFOSTART"], monitorT)
				plg.plugins["EXTERNALINFOEND"] = append(plg.plugins["EXTERNALINFOEND"], monitorT)
				plg.plugins["TRANS_CALL"] = append(plg.plugins["TRANS_CALL"], monitorT)
				plg.plugins["TRANS_CALLCODE"] = append(plg.plugins["TRANS_CALLCODE"], monitorT)
				plg.plugins["TRANS_DELEGATECALL"] = append(plg.plugins["TRANS_DELEGATECALL"], monitorT)
				plg.plugins["TRANS_STATICCALL"] = append(plg.plugins["TRANS_STATICCALL"], monitorT)
			case "handleIAL_MEMORY":
				monitorT := monitorType{} 
				monitorT.mtor  = mtor
				monitorT.option  ="handleIAL_MEMORY"
				plg.plugins["SHA3"] = append(plg.plugins["SHA3"], monitorT)
				plg.plugins["CALLDATACOPY"] = append(plg.plugins["CALLDATACOPY"], monitorT)
				plg.plugins["CODECOPY"] = append(plg.plugins["CODECOPY"], monitorT)
				plg.plugins["RETURNDATACOPY"] = append(plg.plugins["RETURNDATACOPY"], monitorT)
				plg.plugins["MLAOD"] = append(plg.plugins["MLAOD"], monitorT)
				plg.plugins["MSTORE"] = append(plg.plugins["MSTORE"], monitorT)
				plg.plugins["MSTORE8"] = append(plg.plugins["MSTORE8"], monitorT)
				plg.plugins["CREATESTART"] = append(plg.plugins["CREATESTART"], monitorT)
				plg.plugins["CREATEEND"] = append(plg.plugins["CREATEEND"], monitorT)
				plg.plugins["CREATE2START"] = append(plg.plugins["CREATE2START"], monitorT)
				plg.plugins["CREATE2END"] = append(plg.plugins["CREATE2END"], monitorT)
				plg.plugins["CALLSTART"] = append(plg.plugins["CALLSTART"], monitorT)
				plg.plugins["CALLEND"] = append(plg.plugins["CALLEND"], monitorT)
				plg.plugins["CALLCODESTART"] = append(plg.plugins["CALLCODESTART"], monitorT)
				plg.plugins["CALLCODEEND"] = append(plg.plugins["CALLCODEEND"], monitorT)
				plg.plugins["DELEGATECALLSTART"] = append(plg.plugins["DELEGATECALLSTART"], monitorT)
				plg.plugins["DELEGATECALLEND"] = append(plg.plugins["DELEGATECALLEND"], monitorT)
				plg.plugins["STATICCALLSTART"] = append(plg.plugins["STATICCALLSTART"], monitorT)
				plg.plugins["STATICCALLEND"] = append(plg.plugins["STATICCALLEND"], monitorT)
				plg.plugins["RETURN"] = append(plg.plugins["RETURN"], monitorT)
			
			case "handleIAL_STORAGE":
				monitorT := monitorType{} 
				monitorT.mtor  = mtor
				monitorT.option  ="handleIAL_STORAGE"

				plg.plugins["SLOAD"] = append(plg.plugins["SLOAD"], monitorT)
				plg.plugins["SSTORE"] = append(plg.plugins["SSTORE"], monitorT)

			case "handleIAL_ETH":
				monitorT := monitorType{} 
				monitorT.mtor  = mtor
				monitorT.option  ="handleIAL_ETH"
				plg.plugins["TRANS_CREATE"] = append(plg.plugins["TRANS_CREATE"], monitorT)
				plg.plugins["TRANS_CALL"] = append(plg.plugins["TRANS_CALL"], monitorT)
				plg.plugins["TRANS_CALLCODE"] = append(plg.plugins["TRANS_CALLCODE"], monitorT)
				plg.plugins["TRANS_SUICIDE"] = append(plg.plugins["TRANS_SUICIDE"], monitorT)
			
			case "handleIAL_BALANCE":
				monitorT := monitorType{} 
				monitorT.mtor  = mtor
				monitorT.option  ="handleIAL_BALANCE"
				plg.plugins["EXTERNALINFOSTART"] = append(plg.plugins["EXTERNALINFOSTART"], monitorT)
				plg.plugins["EXTERNALINFOEND"] = append(plg.plugins["EXTERNALINFOEND"], monitorT)
				plg.plugins["CALLSTART"] = append(plg.plugins["CALLSTART"], monitorT)
				plg.plugins["CALLEND"] = append(plg.plugins["CALLEND"], monitorT)
				plg.plugins["CALLCODESTART"] = append(plg.plugins["CALLCODESTART"], monitorT)
				plg.plugins["CALLCODEEND"] = append(plg.plugins["CALLCODEEND"], monitorT)
				plg.plugins["CREATESTART"] = append(plg.plugins["CREATESTART"], monitorT)
				plg.plugins["CREATEEND"] = append(plg.plugins["CREATEEND"], monitorT)
				plg.plugins["CREATE2START"] = append(plg.plugins["CREATE2START"], monitorT)
				plg.plugins["CREATE2END"] = append(plg.plugins["CREATE2END"], monitorT)
				plg.plugins["SELFDESTRUCT"] = append(plg.plugins["SELFDESTRUCT"], monitorT)

			case "handleIAL_CONTROLFLOW":
				monitorT := monitorType{} 
				monitorT.mtor  = mtor
				monitorT.option  ="handleIAL_CONTROLFLOW"
				plg.plugins["JUMP"] = append(plg.plugins["JUMP"], monitorT)
				plg.plugins["JUMPI"] = append(plg.plugins["JUMPI"], monitorT)
			
			case "handleIAL_COMPARISON":
				monitorT := monitorType{} 
				monitorT.mtor  = mtor
				monitorT.option  ="handleIAL_COMPARISON"
				plg.plugins["LT"] = append(plg.plugins["LT"], monitorT)
				plg.plugins["GT"] = append(plg.plugins["GT"], monitorT)
				plg.plugins["SLT"] = append(plg.plugins["SLT"], monitorT)
				plg.plugins["SGT"] = append(plg.plugins["SGT"], monitorT)
				plg.plugins["NOT"] = append(plg.plugins["NOT"], monitorT)
				plg.plugins["EQ"] = append(plg.plugins["EQ"], monitorT)
				plg.plugins["ISZERO"] = append(plg.plugins["ISZERO"], monitorT)
			
			case "handleIAL_ARITHMETIC":
				monitorT := monitorType{} 
				monitorT.mtor  = mtor
				monitorT.option  ="handleIAL_ARITHMETIC"
				plg.plugins["ADD"] = append(plg.plugins["ADD"], monitorT)
				plg.plugins["MUL"] = append(plg.plugins["MUL"], monitorT)
				plg.plugins["SUB"] = append(plg.plugins["SUB"], monitorT)
				plg.plugins["DIV"] = append(plg.plugins["DIV"], monitorT)
				plg.plugins["SDIV"] = append(plg.plugins["SDIV"], monitorT)
				plg.plugins["MOD"] = append(plg.plugins["MOD"], monitorT)
				plg.plugins["SMOD"] = append(plg.plugins["SMOD"], monitorT)
				plg.plugins["ADDMOD"] = append(plg.plugins["ADDMOD"], monitorT)
				plg.plugins["MULMOD"] = append(plg.plugins["MULMOD"], monitorT)
				plg.plugins["EXP"] = append(plg.plugins["EXP"], monitorT)
			
			case "handleIAL_EVENT":
				monitorT := monitorType{} 
				monitorT.mtor  = mtor
				monitorT.option  ="handleIAL_EVENT"
				plg.plugins["LOG0"] = append(plg.plugins["LOG0"], monitorT)
				plg.plugins["LOG1"] = append(plg.plugins["LOG1"], monitorT)
				plg.plugins["LOG2"] = append(plg.plugins["LOG2"], monitorT)
				plg.plugins["LOG3"] = append(plg.plugins["LOG3"], monitorT)
				plg.plugins["LOG4"] = append(plg.plugins["LOG4"], monitorT)
			
			default:
				monitorT := monitorType{} 
				monitorT.mtor  = mtor
				monitorT.option  = value
				plg.plugins[value] = append(plg.plugins[value], monitorT)
			}
			// plg.plugins[value] = append(plg.plugins[value], monitorT)
		} else {
			if value == "*" {
				for key, _ := range RetunOpcodeMap() {
					monitorT := monitorType{} 
					monitorT.mtor  = mtor
					monitorT.option  = key
					plg.plugins[key] = append(plg.plugins[key], monitorT)
				}
				break
			}

			plg.regexpString(value, mtor)
		}
	}
}

func (plg *PluginManages) GetOpcodeRegister(opcode string) bool {
	_, isTrue := plg.plugins[opcode]
	return isTrue
}

func (plg *PluginManages) SendDataToPlugin(opcode string, data *collector.CollectorDataT) bool {
	if funcArr, isTrue := plg.plugins[opcode]; isTrue {
		for index := 0; index < len(funcArr); index++ {
			//if opcode == "EQ" {
			//fmt.Println("==========status eq:", funcArr[index].GetStatus())
			//}
			if (funcArr[index].mtor).GetStatus() {
				//close this func
				data.Option = (plg.plugins[opcode])[index].option
				level,comments := ((plg.plugins[opcode])[index].mtor).SendData(data)
				switch level{
				case 0x01:
					StandardWarningReport((funcArr[index].mtor).GetName(),comments,(funcArr[index].mtor).GetLogger(),opcode,2)
					// (plg.plugins[opcode])[index].SetStatus(false)
					// continue
				case 0x02:
					StandardWarningReport((funcArr[index].mtor).GetName(),comments,(funcArr[index].mtor).GetLogger(),opcode,3)
					((plg.plugins[opcode])[index].mtor).SetStatus(false)
					tingrong.BLOCKING_FLAG = true
					continue
				// case 0x03:
				// 	StandardWarningReport(funcArr[index].GetName(),comments,funcArr[index].GetLogger(),3)
				// 	(plg.plugins[opcode])[index].SetStatus(false)
				// 	tingrong.BLOCKING_FLAG = true
				// 	continue
				}

			}
		}
	}

	return true
}

func (plg *PluginManages) Start() {
	for _, valuelist := range plg.plugins {
		for index := 0; index < len(valuelist); index++ {
			(valuelist[index].mtor).SetStatus(true)
		}
	}
}

func (plg *PluginManages) Stop() {
	for _, valuelist := range plg.plugins {
		for index := 0; index < len(valuelist); index++ {
			(valuelist[index].mtor).SetStatus(false)
		}
	}
}

func (plg *PluginManages) SendResultPlugin(sel bool, data *collector.CollectorDataT) {
	opCode := "FAILINFO"
	if sel {
		opCode = "SUCCESSINFO"
	}

	plg.SendDataToPlugin(opCode, data)
}

func StandardWarningReport(PluginName ,comments string,logger *log.ErrTxLog,opcode string,level int){
	txhash := tingrong.TxHash
	var contract string
	if opcode == "EXTERNALINFOSTART" && len(tingrong.CALL_STACK) == 0{
		contract = "EXTERNALCREATE"
	}else{
		temp_str := tingrong.CALL_STACK[len(tingrong.CALL_STACK)-1]
		temp_arr := strings.Split(temp_str,"#")
		contract = temp_arr[0]
	}
	
	logger.CheckIfCreateNewFile()
	logger.OpenFile()
	var logstr string
	if level == 2{
		logstr = txhash + ","+contract+",Warning:"+comments+"\n"
	}else{
		logstr = txhash + ","+contract+",Serious:"+comments+"\n"
	}
	
	logger.WriteLog(logstr)
	logger.CloseFile()

}

//feifei-unreg
func (plg *PluginManages) UnRegisterPlg() {
	for plgkey, valuelist := range plg.plugins {
		for index := 0; index < len(valuelist); index++ {
			//如果valuelist长度为1，就可以删除这个key。否则直接注销是没法注销的
			plgname := (valuelist[index].mtor).GetName()
			if plgname == fei.UnPlg {
				if len(valuelist) == 1 {
					plg.plugins[plgkey] = clearvalue
				} else {
					valuelist = append(valuelist[:index], valuelist[index+1:]...)
					plg.plugins[plgkey] = valuelist
				}

			}
		}
	}
}