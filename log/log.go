package log

import (
	"bufio"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
const LEVEL_TRACE = "TRAC"

const LLEVEL_PANIC = logrus.PanicLevel
const LLEVEL_FATAL = logrus.FatalLevel
const LLEVEL_ERROR = logrus.ErrorLevel
const LLEVEL_WARN = logrus.WarnLevel
const LLEVEL_INFO = logrus.InfoLevel
const LLEVEL_DEBUG = logrus.DebugLevel
const LLEVEL_TRACE = logrus.TraceLevel //like sql trace

var logsAbsFolder string
var logsAllAbsFolder string
var logsErrorAbsFolder string

type Fields = logrus.Fields

var ErrorKey = logrus.ErrorKey

var ShowColor bool

func init() {
	if runtime.GOOS == "windows" {
		ShowColor = false
	} else {
		ShowColor = true
	}
}

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
}

func (logger *LocalLog) ResetLevel(loglevel string) error {

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
	case LEVEL_DEBUG:
		LLevel = logrus.DebugLevel
	case LEVEL_TRACE:
		LLevel = logrus.TraceLevel
	default:
		return errors.New("no such level:" + loglevel)
	}

	alllogfile := filepath.Join(logger.ALL_LogfolderABS, "all_log")
	errlogfile := filepath.Join(logger.ERR_LogfolderABS, "err_log")

	rotateFileHook_ALL, err_all := NewRotateFileHook(RotateFileConfig{
		Filename:   alllogfile,
		MaxSize:    logger.MaxSize, // megabytes
		MaxBackups: logger.MaxBackups,
		MaxAge:     logger.MaxAge, //days
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
		NoColors:        !ShowColor,
	}})

	/////set hooks
	logger.SetLevel(LLevel)
	logger.ReplaceHooks(make(logrus.LevelHooks))
	logger.AddHook(rotateFileHook_ALL)
	logger.AddHook(rotateFileHook_ERR)

	return nil
}

// Default is info level
func New(logsAbsFolder_ string, fileMaxSizeMBytes int, MaxBackupsFiles int, MaxAgeDays int) (*LocalLog, error) {

	logger := logrus.New()

	logsAbsFolder = filepath.Join(logsAbsFolder_, "")
	logsAllAbsFolder = filepath.Join(logsAbsFolder, "all")
	logsErrorAbsFolder = filepath.Join(logsAbsFolder, "error")

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
		fileMaxSizeMBytes, MaxBackupsFiles, MaxAgeDays}
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
		PrintlnColor(Red, err.Error())
		PrintlnColor(White, "exit")
		return
	}
	if len(alllogfiles) == 0 {
		PrintlnColor(White, "no logfile")
		return
	}

	Counter := 0

	for i := 0; i < len(alllogfiles); i++ {
		fname := filepath.Join(folder, alllogfiles[i])

		var cmd *exec.Cmd

		if runtime.GOOS == "windows" {
			cmd = exec.Command("powershell -command " + `"` + " & {Get-Content " + fname + "  | Select-Object -last " + strconv.Itoa(lastN) + " }" + ` "`)
		} else {
			cmd = exec.Command("tail", "-n", strconv.Itoa(lastN), fname)
		}

		stdout, err := cmd.Output()
		if err != nil {
			PrintlnColor(Red, err.Error())
			PrintlnColor(White, "exit")
			return
		}
		lines := splitLines(string(stdout))
		for i := 0; i < len(lines); i++ {

			if strings.Contains(lines[i], "["+LEVEL_DEBUG+"]") {
				PrintlnColor(White, lines[i])
			} else if strings.Contains(lines[i], "["+LEVEL_TRACE+"]") {
				PrintlnColor(Cyan, lines[i])
			} else if strings.Contains(lines[i], "["+LEVEL_INFO+"]") {
				PrintlnColor(Green, lines[i])
			} else if strings.Contains(lines[i], "["+LEVEL_WARN+"]") {
				PrintlnColor(Yellow, lines[i])
			} else if strings.Contains(lines[i], "["+LEVEL_FATAL+"]") ||
				strings.Contains(lines[i], "["+LEVEL_ERROR+"]") ||
				strings.Contains(lines[i], "["+LEVEL_PANIC+"]") {
				PrintlnColor(Red, lines[i])
			} else {
				PrintlnColor(White, lines[i])
			}

			Counter++
			if Counter >= lastN {
				PrintlnColor(White, "END")
				return
			}
		}

	}
	PrintlnColor(White, "EXIT")
}

func splitLines(s string) []string {
	var lines []string
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines
}
