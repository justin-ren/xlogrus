/**
 * @project xlogrus
 * @author justin-ren
 * @desc create log folder, configure logrus with fileHook log rotated, log format and so on
 * @date 5:00 PM 2/9/23
 **/

package common

import (
	"fmt"
	logRotate "github.com/lestrrat-go/file-rotatelogs"
	"github.com/pkg/errors"
	fileLogHook "github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	logFmt "github.com/x-cray/logrus-prefixed-formatter"
	"os"
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
	//"%Y%m%d%H%M%S" for yyyymmddhhmmss
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

func InitOpt() *OptLog {
	return &OptLog{
		StdoutTimeFormat:  "06/01/02 15:04:05",
		LogFileTimeFormat: "2006-01-02 15:04:05.000000",
		//path for all logs
		LogPath:                  "./logs/",
		FileNamePrefix:           "log",
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
	logWriter, err := logRotate.New(fmt.Sprintf("%s.%s", FileNamePrefix, opt.FileNameSuffixTimeFormat),
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
		//全路径日志前缀
		FileNamePrefix = fmt.Sprintf("%s%s", opt.LogPath, opt.ErrLogPrefix)
		//errWriter for error log such as ./logs/error.log.202301
		errWriter, err := logRotate.New(fmt.Sprintf("%v.%v", FileNamePrefix, opt.ErrLogSuffix),
			logRotate.WithLinkName(FileNamePrefix),           // such as ln -s error.log.202301 error.log
			logRotate.WithMaxAge(-1),                         //disable keep log by create time
			logRotate.WithRotationCount(uint(opt.KeepCount)), //keep log by count
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
