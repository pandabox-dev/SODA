package pluginManage

//add new file

import (
	"fmt"
	"github.com/ethereum/collector"
	"os"
	"path/filepath"
	"plugin"
	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type RegisterInfo struct {
	PluginName string   `json:"pluginname"`
	OpCode     map[string]string `json:"option"`
}

func SetUpPlugin(manage *PluginManages){
	pluginFiles,_ := filepath.Glob("./plugin/*.so")
	log_path := "./plugin_log"
	_,err := os.Stat(log_path)
	if err == nil || os.IsNotExist(err){
		os.Mkdir(log_path,os.ModePerm)
	}
	for _, value := range pluginFiles {
		fmt.Println("plugin:", value)
		RegisterPlugin(manage, value)
	}
	
}

func RegisterPlugin(manage *PluginManages, path string) bool {
	plugin, err := plugin.Open(path)
	if err != nil {
		fmt.Println("error open plugin: ", err, "from path :", path)
		os.Exit(-1)
	}
	// fmt.Println("ex",plugin)
	register_method, err := plugin.Lookup("Register")
	if err != nil {
		fmt.Println("Can not find register function:Register() in plugin", err, "from path :", path)
		panic(err)
	}
	register_res, b_err := register_method.(func() []byte)
	if !b_err{
		panic(b_err)
	}
	var register_info RegisterInfo
	err = json.Unmarshal(register_res(), &register_info)
	if err != nil {
		fmt.Println("Can not parse the struct RegisterInfo from the function:Register() in plugin", err, "from path :", path)
		panic(err)
	}
	fmt.Println("Data log path:./plugin_log/" , register_info.PluginName , "datalog")
	register_map := register_info.OpCode
	for opcode,sendfunc := range(register_map){
		var monitor MonitorType
		monitor.SetPluginName(register_info.PluginName)
		monitor.SetLogger(register_info.PluginName)
		// fmt.Println("opcode:",opcode,"sendfunc:",sendfunc)
		symGreeter, err := plugin.Lookup(sendfunc)
		if err != nil {
			fmt.Println("Can not find function",sendfunc," in plugin", err, "from path :", path)
			panic(err)
		}
		rcvefunc, ok := symGreeter.(func(*collector.AllCollector) (byte,string))
		if !ok {
			fmt.Println("unexpected type from module symbol")
			os.Exit(0)
		}
		// fmt.Println("rcve",rcvefunc)
		monitor.SetSendFunc(rcvefunc)
		monitor.SetOpcode(opcode)
		monitor.SetIAL_Optinon(opcode)
		manage.RegisterOpcode(opcode,&monitor)
	}
	return true
}
