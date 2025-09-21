package logger

import (
	// "fmt" "log"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
)

var MsgRecv *log.Logger
var MsgSend *log.Logger

var Err *log.Logger
var Info *log.Logger
var Warn *log.Logger
var Tip *log.Logger
var nullFile *os.File

func init() {
	color.NoColor = false
	MsgRecv = log.New(os.Stdout, color.HiGreenString("<<< "), log.Lmicroseconds|log.Ldate)
	MsgSend = log.New(os.Stdout, color.HiYellowString(">>> "), log.Lmicroseconds|log.Ldate)
	Err = log.New(os.Stderr, color.RedString("error "), log.Ltime|log.Lshortfile)
	Tip = log.New(os.Stderr, color.YellowString("Tip "), log.Ltime|log.Lshortfile)
	Warn = log.New(os.Stderr, color.BlackString("warn "), log.Ltime|log.Lshortfile)
	Info = log.New(os.Stdout, "", log.Lshortfile|log.Ldate|log.Ltime)
	nullFile, _ = os.Create(os.DevNull)

}

var logFile *os.File

func CloseOutputToFile() {
	logFile.Close()
	Err.SetOutput(os.Stdout)
	Tip.SetOutput(os.Stdout)
	Warn.SetOutput(os.Stdout)
	Info.SetOutput(os.Stdout)
}

func SetOutputToFile() {

	appPath := os.Args[0]
	appName := filepath.Base(appPath)

	dir := "log"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			fmt.Println("创建目录失败:", err)
		} else {
			fmt.Println("目录创建成功:", dir)
		}
	} else {
		fmt.Println("目录已存在:", dir)
	}

	appName = strings.TrimSuffix(appName, filepath.Ext(appName)) + ".log"

	timeFormat := time.Now().Format("2006-01-02_15-04-05")
	timeFormat = strings.Replace(timeFormat, ":", "-", -1)
	fileName := "log_" + timeFormat + appName

	oldlogfile := fmt.Sprintf("%s/%s", dir, appName)
	newlogfile := fmt.Sprintf("%s/%s", dir, fileName)
	err := os.Rename(oldlogfile, newlogfile)
	if err != nil {
		fmt.Println("Failed to Rename log file: ", err)
	}
	logfile := fmt.Sprintf("%s/%s", dir, appName)
	//logfile := appName

	filehandle, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Failed to open log file: ", err)
	}
	logFile = filehandle
	Err.SetOutput(logFile)
	Tip.SetOutput(logFile)
	Warn.SetOutput(logFile)
	Info.SetOutput(logFile)
}

func Infof(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	Info.Println(s)
}

func DisableMsg() {
	MsgRecv.SetOutput(nullFile)
	MsgSend.SetOutput(nullFile)

}

func EnableMsg() {
	MsgRecv.SetOutput(os.Stdout)
	MsgSend.SetOutput(os.Stdout)
}

func DisableInfo() {
	Info.SetOutput(nullFile)
	Info.SetOutput(nullFile)
}

func EnableInfo() {
	Info.SetOutput(os.Stdout)
	Info.SetOutput(os.Stdout)
}
