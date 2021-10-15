package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/daqnext/LocalLog/log"
)

/////////////////////

var ExEPath string

func GetPath(relpath string) string {
	return filepath.Join(ExEPath, relpath)
}

func configAbsPath() {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		panic(err.Error())
	}
	runPath, err := filepath.Abs(file)
	if err != nil {
		panic(err.Error())
	}
	index := strings.LastIndex(runPath, string(os.PathSeparator))
	ExEPath = runPath[:index]
}

func main() {
	configAbsPath()

	//default is info level
	llog, err := log.New(GetPath("logs"), 2, 20, 30)
	if err != nil {
		panic(err.Error())
	}

	llog.PrintLastN_AllLogs(100)

	// llog.WithFields(log.Fields{
	// 	"f1": "1",
	// 	"f2": "2",
	// }).Error("Total xxx Error Fileds : %d", 2)

	// llog.WithFields(log.Fields{
	// 	"f1": "1",
	// 	"f2": "2",
	// }).Warn("Total  yy Warn Fileds : %d", 2)

	// llog.ResetLevel(log.LEVEL_DEBUG)

	// llog.WithFields(log.Fields{
	// 	"f1": "1",
	// 	"f2": "2",
	// }).Info("Total zzz Fileds : %d", 2)

	// llog.WithFields(log.Fields{
	// 	"f1": "1",
	// 	"f2": "2",
	// }).Debug("Total Debug Fileds : %d", 2)

	// llog.ResetLevel(log.LEVEL_TRACE)

	// llog.WithFields(log.Fields{
	// 	"f1": "1",
	// 	"f2": "2",
	// }).Debug("Total Debug Fileds : %d", 2)

	// fmt.Println("////////////////////////////////////////////////////////////////////////////////")
	// //all logs include all types :debug ,info ,warning ,error,panic ,fatal
	// llog.PrintLastN_AllLogs(100)
	// fmt.Println("////////////////////////////////////////////////////////////////////////////////")
	// //err logs include all types :,error,panic ,fatal
	// llog.PrintLastN_ErrLogs(100)
	// fmt.Println("////////////////////////////////////////////////////////////////////////////////")

	// llog.Warnln("this is warn ln")
	// llog.Warnf("this is warnf %d", 123)

	// llog.Traceln("this is warn ln")

}
