package pluginManage

//add new file

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	// "time"
	"strconv"
)

type WarnTxLog struct {
	LogFile  *os.File
	FileName string
	FileCount int
	InitFileName string
}

func NewPluginLogger() *WarnTxLog{
	wtlog := WarnTxLog{}
	return &wtlog
}

func (wtlog *WarnTxLog) WriteLog(info string) {
	logger := log.New(wtlog.LogFile, "", 0)
	logger.Output(2, info)
}

func (wtlog *WarnTxLog) CloseFile() {
	wtlog.LogFile.Close()
}

func GetFileSize(filename string) int64 {
	var result int64
	filepath.Walk(filename, func(path string, f os.FileInfo, err error) error {
		result = f.Size()
		return nil
	})
	return result
}

func IsFileExists(filename string) bool{
	_,err := os.Stat(filename)
	if err != nil{
		if os.IsExist(err){
			return true
		}
		return false
	}
	return true

}

func (wtlog *WarnTxLog) CheckIfCreateNewFile() {
	if IsFileExists(wtlog.FileName){
		if GetFileSize(wtlog.FileName) > 300722733 {
			wtlog.FileCount += 1
			filename := wtlog.InitFileName+ strconv.Itoa(wtlog.FileCount)
			f, _ := os.Create(filename)
			wtlog.FileName = filename
			wtlog.LogFile = f
		}
	}
}

func (wtlog *WarnTxLog) OpenFile() {
	logFile, err := os.OpenFile(wtlog.FileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		wtlog.LogFile = nil
		fmt.Println("log file open failed ,err:", err)
	} else {
		wtlog.LogFile = logFile
	}
}

func (wtlog *WarnTxLog) InitialFileLog(filename string) {
	wtlog.LogFile = nil
	wtlog.FileCount = 1
	wtlog.FileName = filename + strconv.Itoa(wtlog.FileCount)
	wtlog.InitFileName = filename
}

