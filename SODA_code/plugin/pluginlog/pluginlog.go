package pluginlog

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	// "time"
	"strconv"
)

type ErrTxLog struct {
	LogFile  *os.File
	FileName string
	FileCount int
	InitFileName string
}

func (el *ErrTxLog) WriteLog(info string) {
	logger := log.New(el.LogFile, "", 0)
	logger.Output(2, info)
}

// func (el *ErrTxLog) OpenFilePath(path string) {
// 	logFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
// 	if err != nil {
// 		f, _ := os.Create(path)
// 		el.LogFile = f
// 	} else {
// 		el.LogFile = logFile
// 	}
// 	el.FileName = path
// }

func (el *ErrTxLog) CloseFile() {
	el.LogFile.Close()
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

func (el *ErrTxLog) CheckIfCreateNewFile() {
	if IsFileExists(el.FileName){
		if GetFileSize(el.FileName) > 300722733 {
			el.FileCount += 1
			filename := el.InitFileName+ strconv.Itoa(el.FileCount)
			f, _ := os.Create(filename)
			el.FileName = filename
			el.LogFile = f
		}
	}
}

func (el *ErrTxLog) OpenFile() {
	logFile, err := os.OpenFile(el.FileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		el.LogFile = nil
		fmt.Println("log file open failed ,err:", err)
	} else {
		el.LogFile = logFile
	}
}

func (el *ErrTxLog) InitialFileLog(filename string) {
	el.LogFile = nil
	el.FileCount = 1
	el.FileName = filename + strconv.Itoa(el.FileCount)
	el.InitFileName = filename
}
