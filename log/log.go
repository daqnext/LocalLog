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

const LEVEL_PANIC_STR = "PANI"
const LEVEL_FATAL_STR = "FATA"
const LEVEL_ERROR_STR = "ERRO"
const LEVEL_WARN_STR = "WARN"
const LEVEL_INFO_STR = "INFO"
const LEVEL_DEBUG_STR = "DEBU"
const LEVEL_TRACE_STR = "TRAC"

const LEVEL_PANIC = logrus.PanicLevel
const LEVEL_FATAL = logrus.FatalLevel
const LEVEL_ERROR = logrus.ErrorLevel
const LEVEL_WARN = logrus.WarnLevel
const LEVEL_INFO = logrus.InfoLevel
const LEVEL_DEBUG = logrus.DebugLevel
const LEVEL_TRACE = logrus.TraceLevel //like sql trace

var logsAbsFolder string
var logsAllAbsFolder string
var logsErrorAbsFolder string

type Fields = logrus.Fields

var ErrorKey = logrus.ErrorKey

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
	MaxSize          int
	MaxBackups       int
	MaxAge           int
	LogrusLevel      logrus.Level
}

func (logger *LocalLog) ResetLevel(loglevel logrus.Level) error {

	alllogfile := logger.ALL_LogfolderABS + "/all_log"
	errlogfile := logger.ERR_LogfolderABS + "/err_log"

	rotateFileHook_ALL, err_all := NewRotateFileHook(RotateFileConfig{
		Filename:   alllogfile,
		MaxSize:    logger.MaxSize, // megabytes
		MaxBackups: logger.MaxBackups,
		MaxAge:     logger.MaxAge, //days
		Level:      loglevel,
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
		MaxSize:    logger.MaxSize, // megabytes
		MaxBackups: logger.MaxBackups,
		MaxAge:     logger.MaxAge, //days
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

	logger.SetFormatter(UTCFormatter{&nested.Formatter{
		HideKeys:        false,
		TimestampFormat: "2006-01-02 15:04:05",
	}})

	/////set hooks

	//levelHooks[LLevel] = append(levelHooks[LLevel], rotateFileHook_ALL)
	//levelHooks[logrus.ErrorLevel] = append(levelHooks[logrus.ErrorLevel], rotateFileHook_ERR)
	logger.SetLevel(loglevel)
	logger.ReplaceHooks(make(logrus.LevelHooks))
	logger.AddHook(rotateFileHook_ALL)
	logger.AddHook(rotateFileHook_ERR)

	return nil
}

// Default is info level
func New(logsAbsFolder_ string, fileMaxSizeMBytes int, MaxBackupsFiles int, MaxAgeDays int) (*LocalLog, error) {

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
	//default info level//
	LocalLogPointer := &LocalLog{*logger, logsAllAbsFolder, logsErrorAbsFolder,
		fileMaxSizeMBytes, MaxBackupsFiles, MaxAgeDays, logrus.InfoLevel}
	LocalLogPointer.ResetLevel(LEVEL_INFO)
	return LocalLogPointer, nil
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

			if strings.Contains(lines[i], "["+LEVEL_DEBUG_STR+"]") {
				fmt.Println(string(White), lines[i])
			} else if strings.Contains(lines[i], "["+LEVEL_TRACE_STR+"]") {
				fmt.Println(string(Cyan), lines[i])
			} else if strings.Contains(lines[i], "["+LEVEL_INFO_STR+"]") {
				fmt.Println(string(Green), lines[i])
			} else if strings.Contains(lines[i], "["+LEVEL_WARN_STR+"]") {
				fmt.Println(string(Yellow), lines[i])
			} else if strings.Contains(lines[i], "["+LEVEL_FATAL_STR+"]") ||
				strings.Contains(lines[i], "["+LEVEL_ERROR_STR+"]") ||
				strings.Contains(lines[i], "["+LEVEL_PANIC_STR+"]") {
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
