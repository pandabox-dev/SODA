package pluginManage

//add new file

import (
	"fmt"
	"github.com/ethereum/collector"
	"github.com/ethereum/go-ethereum/cmd/pluginManage/Monitor"
	"os"
	"path/filepath"
	"plugin"
	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type RunOpCode struct {
	MethodName string   `json:"methodname"`
	OpCode     []string `json:"option"`
}

func StartRun(manage *PluginManages) {
	files, _ := filepath.Glob("./plugin/*.so")

	fmt.Println("file_len1ss:", len(files))

	temp_path1 := "./plugin_log"
	_,err_1 := os.Stat(temp_path1)
	// fmt.Println("err:",err_1)
	if err_1 == nil || os.IsNotExist(err_1){
		os.Mkdir(temp_path1,os.ModePerm)
	}

	for _, value := range files {
		fmt.Println("file_name:", value)
		RegisterMethod(manage, value)
	}
}

func RegisterMethod(manage *PluginManages, path string) bool {
	pl, err := plugin.Open(path)
	if err != nil {
		fmt.Println("error open plugin: ", err, "from path :", path)
		os.Exit(-1)
	}
	run, err := pl.Lookup("Run")
	if err != nil {
		panic(err)
	}

	res, y := run.(func() []byte)
	if !y {
		panic(y)
	}

	var rundata RunOpCode
	err = json.Unmarshal(res(), &rundata)
	if err != nil {
		panic(err)
	}
	var method monitor.MonitorMethod
	method.SetMethodName(rundata.MethodName)

	//add new 
	fmt.Println("SetLogger")
	method.SetLogger(rundata.MethodName)
	//add new 

	symGreeter, err := pl.Lookup("Recv")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	rcve, ok := symGreeter.(func(*collector.CollectorDataT) (byte,string))
	fmt.Println("rcve",rcve)
	if !ok {
		fmt.Println("unexpected type from module symbol")
		os.Exit(0)
	}

	method.SetSendDataFunc(rcve)
	manage.RegisterOption(rundata.OpCode, &method)

	return true
}
