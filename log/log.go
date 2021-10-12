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

const LEVEL_FATAL = "FATA"
const LEVEL_PANIC = "PANI"
const LEVEL_ERROR = "ERRO"
const LEVEL_WARN = "WARN"
const LEVEL_INFO = "INFO"
const LEVEL_DEBUG = "DEBU"

var logsAbsFolder string
var logsAllAbsFolder string
var logsErrorAbsFolder string

type Fields map[string]interface{}

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

func Initialize(logsAbsFolder_ string, fileMaxSizeMBytes int, MaxBackupsFiles int, MaxAgeDays int, loglevel string) error {

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

	logsAbsFolder = strings.Trim(logsAbsFolder_, "/")
	logsAllAbsFolder = logsAbsFolder + "/all"
	logsErrorAbsFolder = logsAbsFolder + "/error"

	//make sure the logs folder exist otherwise create dir
	dirError := checkAndMkDir(logsAbsFolder)
	if dirError != nil {
		return dirError
	}
	dirError = checkAndMkDir(logsAllAbsFolder)
	if dirError != nil {
		return dirError
	}
	dirError = checkAndMkDir(logsErrorAbsFolder)
	if dirError != nil {
		return dirError
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
		return err_all
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
		return err_err
	}

	logrus.SetFormatter(UTCFormatter{&nested.Formatter{
		HideKeys:        false,
		TimestampFormat: "2006-01-02 15:04:05",
	}})

	logrus.SetLevel(LLevel)
	logrus.AddHook(rotateFileHook_ALL)
	logrus.AddHook(rotateFileHook_ERR)

	return nil

}

func WithFields(keymap map[string]interface{}) *logrus.Entry {
	return logrus.WithFields(logrus.Fields(keymap))
}

func getLogFilesList(log_folder string) ([]string, error) {

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

func PrintLastN_ErrLogs(lastN int) {
	printLastNLogs("error", lastN)
}

func PrintLastN_AllLogs(lastN int) {
	printLastNLogs("all", lastN)
}

func printLastNLogs(type_ string, lastN int) {

	var alllogfiles []string
	var err error
	var folder string
	if type_ == "error" {
		folder = logsErrorAbsFolder
	} else {
		folder = logsAllAbsFolder
	}
	alllogfiles, err = getLogFilesList(folder)

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
