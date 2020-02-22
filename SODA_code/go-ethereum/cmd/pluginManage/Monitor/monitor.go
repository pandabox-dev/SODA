package monitor

import (
	"github.com/ethereum/collector"
	"github.com/ethereum/go-ethereum/log"
	"os"
	"fmt"
)


//add new file

type SendDataType func(*collector.CollectorDataT) (byte,string)

type MonitorMethod struct {
	status       bool
	methodName   string
	sendDataFunc SendDataType
	logger		 *log.ErrTxLog
}

func (m *MonitorMethod) SetSendDataFunc(sendfunc SendDataType) {
	m.sendDataFunc = sendfunc
}

func (m *MonitorMethod) SendData(data *collector.CollectorDataT) (byte,string) {
	return m.sendDataFunc(data)
}

func (m *MonitorMethod) SetMethodName(name string) {
	m.methodName = name
}

func (m *MonitorMethod) SetStatus(status bool) {
	m.status = status
}

func (m *MonitorMethod) GetStatus() bool {
	return m.status
}

func (m *MonitorMethod) GetName() string {
	return m.methodName
}


func (m *MonitorMethod) SetLogger(FileName string) {
	m.logger = log.NewPluginLogger()
	filepath := "./plugin_log/" + FileName + "datalog/" +FileName+"datalog" 
	m.logger.InitialFileLog(filepath)


	temp_path1 := "./plugin_log/" + FileName + "datalog"
	fmt.Println(temp_path1)
	_,err_1 := os.Stat(temp_path1)
	fmt.Println(err_1)
	if err_1 == nil || os.IsNotExist(err_1){
		os.Mkdir(temp_path1,os.ModePerm)
	}


} 

func (m *MonitorMethod) GetLogger() *log.ErrTxLog {
	return m.logger
}
