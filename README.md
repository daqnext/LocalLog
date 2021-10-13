# LocalLog

### install 
```
go get "github.com/daqnext/LocalLog/log"
```

```go
 package main

import (
	"fmt"

	"github.com/daqnext/LocalLog/log"
)

/////////////////////

func init() {
	//logsAbsFolder_, fileMaxSizeMBytes, MaxBackupsFiles, MaxAgeDays, loglevel
	log.Initialize("logs", 2, 20, 30, log.LEVEL_DEBUG)
}

func main() {

	log.WithFields(log.Fields{
		"f1": "1",
		"f2": "2",
	}).Error("Total Error Fileds : %d", 2)

	log.WithFields(log.Fields{
		"f1": "1",
		"f2": "2",
	}).Warn("Total Warn Fileds : %d", 2)

	log.WithFields(log.Fields{
		"f1": "1",
		"f2": "2",
	}).Info("Total Fileds : %d", 2)

	log.WithFields(log.Fields{
		"f1": "1",
		"f2": "2",
	}).Debug("Total Debug Fileds : %d", 2)

	fmt.Println("////////////////////////////////////////////////////////////////////////////////")
	//all logs include all types :debug ,info ,warning ,error,panic ,fatal
	log.PrintLastN_AllLogs(100)
	fmt.Println("////////////////////////////////////////////////////////////////////////////////")
	//err logs include all types :,error,panic ,fatal
	log.PrintLastN_ErrLogs(100)
	fmt.Println("////////////////////////////////////////////////////////////////////////////////")

	log.Warnln("this is warn ln")
	log.Warnf("this is warnf %d", 123)

}


```
