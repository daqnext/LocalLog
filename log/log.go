package log

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

const LEVEL_PANIC = "PANI"
const LEVEL_FATAL = "FATA"
const LEVEL_ERROR = "ERRO"
const LEVEL_WARN = "WARN"
const LEVEL_INFO = "INFO"
const LEVEL_DEBUG = "DEBU"

var logsAbsFolder string
var logsAllAbsFolder string
var logsErrorAbsFolder string

type Fields = logrus.Fields

func checkAndMkDir(logsAbsFolder string) error {

	info, err := os.Stat(logsAbsFolder)
	if err != nil {
		if os.IsNotExist(err) {
			mkdirerr := os.Mkdir(logsAbsFolder, 0777)
			if mkdirerr != nil {
				return mkdirerr
			} else {
				return nil
			}
		} else {
			return err
		}
	} else {
		if !info.IsDir() {
			mkdirerr := os.Mkdir(logsAbsFolder, 0777)
			if mkdirerr != nil {
				return mkdirerr
			} else {
				return nil
			}
		} else {
			return nil
		}
	}
}

type LocalLog struct {
	logrus.Logger
	ALL_LogfolderABS string
	ERR_LogfolderABS string
}

func New(logsAbsFolder_ string, fileMaxSizeMBytes int, MaxBackupsFiles int, MaxAgeDays int, loglevel string) (*LocalLog, error) {

	var LLevel logrus.Level

	switch loglevel {
	case LEVEL_PANIC:
		LLevel = logrus.PanicLevel
	case LEVEL_FATAL:
		LLevel = logrus.FatalLevel
	case LEVEL_ERROR:
		LLevel = logrus.ErrorLevel
	case LEVEL_WARN:
		LLevel = logrus.WarnLevel
	case LEVEL_INFO:
		LLevel = logrus.InfoLevel
	default:
		LLevel = logrus.DebugLevel
	}

	logger := logrus.New()

	logsAbsFolder = strings.TrimRight(logsAbsFolder_, "/")
	logsAllAbsFolder = logsAbsFolder + "/all"
	logsErrorAbsFolder = logsAbsFolder + "/error"

	//make sure the logs folder exist otherwise create dir
	dirError := checkAndMkDir(logsAbsFolder)
	if dirError != nil {
		return nil, dirError
	}
	dirError = checkAndMkDir(logsAllAbsFolder)
	if dirError != nil {
		return nil, dirError
	}
	dirError = checkAndMkDir(logsErrorAbsFolder)
	if dirError != nil {
		return nil, dirError
	}
	///////////////////////

	alllogfile := logsAllAbsFolder + "/all_log"
	errlogfile := logsErrorAbsFolder + "/err_log"

	rotateFileHook_ALL, err_all := NewRotateFileHook(RotateFileConfig{
		Filename:   alllogfile,
		MaxSize:    fileMaxSizeMBytes, // megabytes
		MaxBackups: MaxBackupsFiles,
		MaxAge:     MaxAgeDays, //days
		Level:      LLevel,
		Formatter: UTCFormatter{&nested.Formatter{
			NoColors:        true,
			HideKeys:        false,
			TimestampFormat: "2006-01-02 15:04:05",
		}},
	})
	if err_all != nil {
		return nil, err_all
	}

	rotateFileHook_ERR, err_err := NewRotateFileHook(RotateFileConfig{
		Filename:   errlogfile,
		MaxSize:    fileMaxSizeMBytes, // megabytes
		MaxBackups: MaxBackupsFiles,
		MaxAge:     MaxAgeDays, //days
		Level:      logrus.ErrorLevel,
		Formatter: UTCFormatter{&nested.Formatter{
			NoColors:        true,
			HideKeys:        false,
			TimestampFormat: "2006-01-02 15:04:05",
		}},
	})

	if err_err != nil {
		return nil, err_err
	}

	logger.SetFormatter(UTCFormatter{&nested.Formatter{
		HideKeys:        false,
		TimestampFormat: "2006-01-02 15:04:05",
	}})

	logger.SetLevel(LLevel)
	logger.AddHook(rotateFileHook_ALL)
	logger.AddHook(rotateFileHook_ERR)

	return &LocalLog{*logger, logsAllAbsFolder, logsErrorAbsFolder}, nil

}

func (logger *LocalLog) GetLogFilesList(log_folder string) ([]string, error) {

	var result []string
	files, err := ioutil.ReadDir(log_folder)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		result = append(result, f.Name())
	}
	return result, nil
}

func (logger *LocalLog) PrintLastN_ErrLogs(lastN int) {
	logger.printLastNLogs("error", lastN)
}

func (logger *LocalLog) PrintLastN_AllLogs(lastN int) {
	logger.printLastNLogs("all", lastN)
}

func (logger *LocalLog) printLastNLogs(type_ string, lastN int) {

	var alllogfiles []string
	var err error
	var folder string
	if type_ == "error" {
		folder = logger.ERR_LogfolderABS
	} else {
		folder = logger.ALL_LogfolderABS
	}
	alllogfiles, err = logger.GetLogFilesList(folder)

	if err != nil {
		fmt.Println(string(Red), err.Error())
		fmt.Println(string(White), "exit")
		return
	}
	if len(alllogfiles) == 0 {
		fmt.Println("no logfile")
		return
	}

	Counter := 0

	for i := 0; i < len(alllogfiles); i++ {
		fname := folder + "/" + alllogfiles[i]
		cmd := exec.Command("tail", "-n", strconv.Itoa(lastN), fname)
		stdout, err := cmd.Output()
		if err != nil {
			fmt.Println(string(Red), err.Error())
			fmt.Println(string(White), "exit")
			return
		}
		lines := splitLines(string(stdout))
		for i := 0; i < len(lines); i++ {

			if strings.Contains(lines[i], "["+LEVEL_DEBUG+"]") {
				fmt.Println(string(White), lines[i])
			} else if strings.Contains(lines[i], "["+LEVEL_INFO+"]") {
				fmt.Println(string(Green), lines[i])
			} else if strings.Contains(lines[i], "["+LEVEL_WARN+"]") {
				fmt.Println(string(Yellow), lines[i])
			} else if strings.Contains(lines[i], "["+LEVEL_FATAL+"]") ||
				strings.Contains(lines[i], "["+LEVEL_ERROR+"]") ||
				strings.Contains(lines[i], "["+LEVEL_PANIC+"]") {
				fmt.Println(string(Red), lines[i])
			} else {
				fmt.Println(string(White), lines[i])
			}

			Counter++
			if Counter >= lastN {
				fmt.Println(string(White), "END")
				return
			}
		}

	}

	fmt.Println(string(White), "EXIT")
}

func splitLines(s string) []string {
	var lines []string
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines
}
