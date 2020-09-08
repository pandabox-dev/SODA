package pluginManage

//add new file

import (
	"github.com/ethereum/collector"
	"strings"
	// "fmt"
	"github.com/ethereum/go-ethereum/tingrong"
	"github.com/ethereum/go-ethereum/fei"
)

//2019.03.01 version plugin

type PluginManages struct {
	plugins map[string][]*MonitorType
	
}

var clearvalue []*MonitorType

func NewPluginManages() *PluginManages {
	return &PluginManages{make(map[string][]*MonitorType)}
}

func (plg *PluginManages) RegisterOpcode(opcode string ,monitor *MonitorType){
	res := IsOpExist(opcode)
	// fmt.Println("res:",res)
	switch res{
	case 1:
		monitor.SetStatus(false)
		plg.plugins[opcode] = append(plg.plugins[opcode], monitor)
	case 2:
		registerIALOp := ReturnIALArray(opcode)
		for _,value := range(registerIALOp){
			// fmt.Println("value:",value)
			monitor.SetStatus(false)
			monitor.SetOpcode(value)
			plg.plugins[value] = append(plg.plugins[value], monitor)
		}
	default:
		if opcode == "*"{
			for key, _ := range RetunOpcodeMap() {
				monitor.SetStatus(false)
				plg.plugins[key] = append(plg.plugins[key], monitor)
			}
			break
		}
	}

}

func (plg *PluginManages) GetOpcodeRegister(opcode string) bool {
	_, isTrue := plg.plugins[opcode]
	return isTrue
}



func (plg *PluginManages) SendDataToPlugin(opcode string, data *collector.AllCollector) bool {
	// if tingrong.TxHash == "0x847194c9081008ede0ca7dbbb037408a15b6b96b11bca07f032af001c2edd083" || tingrong.TxHash == "0x1fa290fac8231ff6936ae22b2d6116ecf7dfe5cda6823ce44cd803ef620aab84"{
	// 	fmt.Println("tingrong.TxHash :",tingrong.TxHash)
	if monitor_arr, isTrue := plg.plugins[opcode]; isTrue {
		for index := 0; index < len(monitor_arr); index++ {
			// true_opcode :=  plg.plugins[opcode][index].GetIAL_Optinon()
			if plg.plugins[opcode][index].GetStatus(){
				
				// fmt.Println("senddata:",data)
				// fmt.Println("new:", plg.plugins[opcode][index])
				warning_level,results := ((plg.plugins[opcode])[index]).Send(data)
				switch warning_level{
				case 0x01:
					StandardWarningReport(((plg.plugins[opcode])[index]).GetPluginName(),results,((plg.plugins[opcode])[index]).GetLogger(),opcode,2)
				case 0x02:
					StandardWarningReport(((plg.plugins[opcode])[index]).GetPluginName(),results,((plg.plugins[opcode])[index]).GetLogger(),opcode,3)
					((plg.plugins[opcode])[index]).SetStatus(false)
					tingrong.BLOCKING_FLAG = true
					continue
				case 0x03:
					StandardWarningReport(((plg.plugins[opcode])[index]).GetPluginName(),results,((plg.plugins[opcode])[index]).GetLogger(),opcode,3)
					((plg.plugins[opcode])[index]).SetStatus(false)
					tingrong.BLOCKING_FLAG = true
					continue
				default:
					continue
				}
				
				
			}
		}
	}else{
		return false
	}
	// }

	return true
}


func (plg *PluginManages) Start() {
	for _, valuelist := range plg.plugins {
		for index := 0; index < len(valuelist); index++ {
			(valuelist[index]).SetStatus(true)
		}
	}
}

func (plg *PluginManages) Stop() {
	for _, valuelist := range plg.plugins {
		for index := 0; index < len(valuelist); index++ {
			(valuelist[index]).SetStatus(false)
		}
	}
}

func StandardWarningReport(PluginName ,comments string,logger *WarnTxLog,opcode string,level int){
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
			plgname := (valuelist[index]).GetPluginName()
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