package main

import (
	"fmt"

	"github.com/daqnext/LocalLog/log"
)

/////////////////////

func main() {

	llog, err := log.New("logs", 2, 20, 30, log.LEVEL_DEBUG)
	if err != nil {
		panic(err.Error())
	}

	llog.WithFields(log.Fields{
		"f1": "1",
		"f2": "2",
	}).Error("Total Error Fileds : %d", 2)

	llog.WithFields(log.Fields{
		"f1": "1",
		"f2": "2",
	}).Warn("Total Warn Fileds : %d", 2)

	llog.WithFields(log.Fields{
		"f1": "1",
		"f2": "2",
	}).Info("Total Fileds : %d", 2)

	llog.WithFields(log.Fields{
		"f1": "1",
		"f2": "2",
	}).Debug("Total Debug Fileds : %d", 2)

	fmt.Println("////////////////////////////////////////////////////////////////////////////////")
	//all logs include all types :debug ,info ,warning ,error,panic ,fatal
	llog.PrintLastN_AllLogs(100)
	fmt.Println("////////////////////////////////////////////////////////////////////////////////")
	//err logs include all types :,error,panic ,fatal
	llog.PrintLastN_ErrLogs(100)
	fmt.Println("////////////////////////////////////////////////////////////////////////////////")

	llog.Warnln("this is warn ln")
	llog.Warnf("this is warnf %d", 123)

}
