/*
 * @Author: justin-ren
 * @Date: 2025-02-26 02:11:15
 * @LastEditors: justin-ren
 * @LastEditTime: 2025-03-04 09:10:35
 * @FilePath: /xlogrus-edit/xlogrus/common/common.go
 * @Description: basic configuration for logrus
 *
 */

package common

import (
	"fmt"
	"os"

	logRotate "github.com/lestrrat-go/file-rotatelogs"
	"github.com/pkg/errors"
	fileLogHook "github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	logFmt "github.com/x-cray/logrus-prefixed-formatter"
)

type OptLog struct {

	//time format for screen log
	StdoutTimeFormat string
	//time format for log file
	LogFileTimeFormat string
	//path for all logs
	LogPath string
	//example prefix 'access' in access.log.20230105
	FileNamePrefix string
	//example suffix '%Y%m%d' for '20230105' in access.log.20230105
	FileNameSuffixTimeFormat string
	//error level log will be saved in seperated error.log for keeping more time if it's ture
	SetErrFileHook bool
	//keep log count
	KeepCount int
	//log level
	LogLevel logrus.Level
	//logFile
	//MapLogFile map[string]string
	//error log prefix error.log if SetErrFileHook is true
	ErrLogPrefix string
	//error log suffix "%Y%m" if SetErrFileHook is true
	ErrLogSuffix string
}

// SetStdoutTimeFormat sets the stdout time format.
func (o *OptLog) SetStdoutTimeFormat(format string) error {
	o.StdoutTimeFormat = format
	return nil
}

// SetLogFileTimeFormat sets the log file time format.
func (o *OptLog) SetLogFileTimeFormat(format string) error {
	o.LogFileTimeFormat = format
	return nil
}

// SetLogPath sets the path for all logs.
func (o *OptLog) SetLogPath(path string) error {
	if path == "" {
		return errors.New("log path cannot be empty")
	}
	o.LogPath = path
	return nil
}

// SetFileNamePrefix sets the filename prefix.
func (o *OptLog) SetFileNamePrefix(prefix string) error {
	o.FileNamePrefix = prefix
	return nil
}

// SetFileNameSuffixTimeFormat sets the filename suffix time format.
func (o *OptLog) SetFileNameSuffixTimeFormat(format string) error {
	o.FileNameSuffixTimeFormat = format
	return nil
}

// SetSetErrFileHook sets whether to create a separate error log file.
func (o *OptLog) SetSetErrFileHook(enabled bool) error {
	o.SetErrFileHook = enabled
	return nil
}

// SetKeepCount sets the number of logs to keep.
func (o *OptLog) SetKeepCount(count int) error {
	if count < 0 {
		return errors.New("keep count cannot be negative")
	}
	o.KeepCount = count
	return nil
}

// SetLogLevel sets the log level.
// level : trace,debug,warn,warning,error,fatal,panic
func (o *OptLog) SetLogLevel(level string) error {
	if lvl, err := logrus.ParseLevel(level); err != nil {
		return err
	} else {
		o.LogLevel = lvl
	}
	return nil
}

// SetErrLogPrefix sets the error log prefix.
func (o *OptLog) SetErrLogPrefix(prefix string) error {
	o.ErrLogPrefix = prefix
	return nil
}

// SetErrLogSuffix sets the error log suffix.
func (o *OptLog) SetErrLogSuffix(suffix string) error {
	o.ErrLogSuffix = suffix
	return nil
}

/*InitOpt
 * @msg init logrus params
 * @return: *OptLog
 */
func InitOpt() *OptLog {
	return &OptLog{
		StdoutTimeFormat:  "06/01/02 15:04:05",
		LogFileTimeFormat: "2006-01-02 15:04:05.000000",
		//path for all logs
		LogPath:        "./logs/",
		FileNamePrefix: "log", //log in access.log.20230105
		//not work if specify %H%m%s, but below is enough
		FileNameSuffixTimeFormat: "%Y%m%d",
		SetErrFileHook:           true,
		//keep log count
		KeepCount: 7,
		//log level
		LogLevel:     logrus.DebugLevel,
		ErrLogPrefix: "error.log",
		ErrLogSuffix: "%Y%m",
	}
}

/**
 * @description: LogOption为接受With函数的接口，
 *			T为OptLog结构体，Apply为OptLog赋值的Set函数
 */
type LogOption[T any] interface {
	Apply(*T) error
}

/**
 * @description: 实现LogOption接口的函数的类型,对应不同结构体的Set函数
 * @param {} T LogOpt的地址
 * @return {*}
 */
type LogOptionFunc[T any] func(*T) error

/**
 * @description: 实现接口LogOption的Apply函数，
 *				调用LogOpt类的结构体的Set函数，修改结构体默认值opt
 * @param {*T} opt 有默认值的LogOpt结构指针
 */
func (f LogOptionFunc[T]) Apply(opt *T) error {
	return f(opt)
}

func NewLogOptionFunc[T any](fn func(*T) error) LogOption[T] {
	return LogOptionFunc[T](fn)
}

/*ConfigLogrus
 * @msg to configure logrus with
 * 		1. log format with color, timestamp
 *	   	2. log redirect by loglevel
 *		3. create file link for alive log
 *		4. set log keeping count
 * @receiver opt
 * @return: *logrus.Logger
 * @return: error
 */
func (opt *OptLog) ConfigLogrus() (*logrus.Logger, error) {
	log := logrus.New()
	log.SetLevel(opt.LogLevel)
	//set log format for standard output
	stdoutFmt := &logFmt.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: opt.StdoutTimeFormat, //timestamp for standard output
		ForceFormatting: true,
		ForceColors:     true,
		DisableColors:   false,
	}
	log.SetFormatter(stdoutFmt)

	logFileFmt := &logFmt.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: opt.LogFileTimeFormat, //timestamp for log file
		ForceFormatting: true,
		ForceColors:     false,
		DisableColors:   true,
	}

	if err := os.MkdirAll(opt.LogPath, 0775); err != nil {
		return log, errors.Cause(err)
	}
	FileNamePrefix := fmt.Sprintf("%s%s", opt.LogPath, opt.FileNamePrefix)
	logWriter, err := logRotate.New(fmt.Sprintf("%v.%v", FileNamePrefix, opt.FileNameSuffixTimeFormat),
		logRotate.WithLinkName(FileNamePrefix),           //create log link, such as ln -s access.log.20230205 access.log
		logRotate.WithMaxAge(-1),                         //disable remove log by create time
		logRotate.WithRotationCount(uint(opt.KeepCount)), //set count for keeping log
	)
	if err != nil {
		return log, errors.Cause(err)
	}

	logHook := fileLogHook.NewHook(fileLogHook.WriterMap{
		logrus.DebugLevel: logWriter,
		logrus.InfoLevel:  logWriter,
		logrus.WarnLevel:  logWriter,
		logrus.ErrorLevel: logWriter,
		logrus.FatalLevel: logWriter,
	}, logFileFmt)
	//opt.MapLogFile[opt.FileNamePrefix] = fmt.Sprintf("%s.%s", FileNamePrefix, opt.FileNameSuffixTimeFormat)
	//writing log to file when printing to screen by hook
	log.AddHook(logHook)
	fmt.Println(logWriter.CurrentFileName())
	if opt.SetErrFileHook {
		//full path for log filename
		FileNamePrefix = fmt.Sprintf("%s%s", opt.LogPath, opt.ErrLogPrefix)
		//errWriter for error log such as ./logs/error.log.202301
		errWriter, err := logRotate.New(fmt.Sprintf("%v.%v", FileNamePrefix, opt.ErrLogSuffix),
			logRotate.WithLinkName(FileNamePrefix),           // such as ln -s error.log.202301 error.log
			logRotate.WithMaxAge(-1),                         //disable keep log by created time
			logRotate.WithRotationCount(uint(opt.KeepCount)), //keep alive log by count
		)

		if err != nil {
			return log, errors.Cause(err)
		}
		//add hook for error.log with level warn,error, fatal
		errHook := fileLogHook.NewHook(fileLogHook.WriterMap{
			logrus.WarnLevel:  errWriter,
			logrus.ErrorLevel: errWriter,
			logrus.FatalLevel: errWriter,
		}, logFileFmt)

		log.AddHook(errHook)
		//opt.MapLogFile = make(map[string]string)

		//opt.MapLogFile[opt.ErrLogPrefix] = errWriter.CurrentFileName()
	}
	return log, nil
}
