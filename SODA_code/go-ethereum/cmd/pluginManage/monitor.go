package pluginManage

import (
	"github.com/ethereum/collector"
	"os"
	// "fmt"
)


//add new file

type SendFuncType func(*collector.AllCollector) (byte,string)

type MonitorType struct {
	Status 		bool
	SendFunc 	SendFuncType
	Opcode 		string
	Logger 		*WarnTxLog
	IAL_Optinon	string
	PluginName 	string
}

func (m *MonitorType) SetStatus(Status bool) {
	m.Status = Status
}
func (m *MonitorType) GetStatus() bool {
	return m.Status
}

func (m *MonitorType) SetSendFunc(SendFunc SendFuncType) {
	m.SendFunc = SendFunc
}
func (m *MonitorType) GetSendFunc() SendFuncType {
	return m.SendFunc
}
func (m *MonitorType) Send(data *collector.AllCollector) (byte,string) {
	return m.SendFunc(data)
}

func (m *MonitorType) SetOpcode(Opcode string) {
	m.Opcode = Opcode
}
func (m *MonitorType) GetOpcode() string {
	return m.Opcode
}

func (m *MonitorType) SetIAL_Optinon(IAL_Optinon string) {
	m.IAL_Optinon = IAL_Optinon
}
func (m *MonitorType) GetIAL_Optinon() string {
	return m.IAL_Optinon
}

func (m *MonitorType) SetPluginName(PluginName string) {
	m.PluginName = PluginName
}
func (m *MonitorType) GetPluginName() string {
	return m.PluginName
}


func (m *MonitorType) SetLogger(FileName string) {
	m.Logger = NewPluginLogger()
	filepath := "./plugin_log/" + FileName + "datalog/" +FileName+"datalog" 
	m.Logger.InitialFileLog(filepath)

	logpath := "./plugin_log/" + FileName + "datalog"
	// fmt.Println("Data log path:",logpath)
	_,err_1 := os.Stat(logpath)
	// fmt.Println(err_1)
	// fmt.Println(os.IsNotExist(err_1))
	if err_1 == nil || os.IsNotExist(err_1){
		os.Mkdir(logpath,os.ModePerm)
	}


} 

func (m *MonitorType) GetLogger() *WarnTxLog {
	return m.Logger
}
