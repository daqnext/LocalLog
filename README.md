# LocalLog

##### support both linux mac andwindows

### install 
```
go get "github.com/daqnext/LocalLog/log"
```

### run test
```bash

go build 

./LocalLog   //for linux mac
./LocalLog.exe //for windows

```

```go

package main

import (
	"fmt"

	"github.com/daqnext/LocalLog/log"
	"github.com/daqnext/utils/path"
)

func main() {

	//default is info level
	llog, err := log.New(path.GetAbsPath("logs"), 2, 20, 30)
	if err != nil {
		panic(err.Error())
	}

	llog.WithFields(log.Fields{
		"f1": "1",
		"f2": "2",
	}).Error("Total xxx Error Fileds : %d", 2)

	llog.WithFields(log.Fields{
		"f1": "1",
		"f2": "2",
	}).Warn("Total  yy Warn Fileds : %d", 2)

	llog.ResetLevel(log.LEVEL_DEBUG)

	llog.WithFields(log.Fields{
		"f1": "1",
		"f2": "2",
	}).Info("Total zzz Fileds : %d", 2)

	llog.WithFields(log.Fields{
		"f1": "1",
		"f2": "2",
	}).Debug("Total Debug Fileds : %d", 2)

	llog.ResetLevel(log.LEVEL_TRACE)

	fmt.Println("////////////////////////////////////////////////////////////////////////////////")
	//all logs include all types :debug ,info ,warning ,error,panic ,fatal
	llog.PrintLastN_AllLogs(100)
	fmt.Println("////////////////////////////////////////////////////////////////////////////////")
	//err logs include all types :,error,panic ,fatal
	llog.PrintLastN_ErrLogs(100)
	fmt.Println("////////////////////////////////////////////////////////////////////////////////")

	llog.Warnln("this is warn ln")
	llog.Warnf("this is warnf %d", 123)

	llog.Traceln("this is warn ln")

}


```
