/**
 * @project xlogrus
 * @author justin-ren
 * @desc //TODO
 * @date 5:00 PM 2/9/23
 **/

package common

import (
	"github.com/sirupsen/logrus"
)

type OptLog struct {

	//time format for screen adapter
	StdoutTimeFormat string
	//time format for adapter file
	LogFileTimeFormat string
	//path for all logs
	LogPath string
	//example prefix 'access' in access.adapter.20230105
	FileNamePrefix string
	//example suffix '%Y%m%d' for '20230105' in access.adapter.20230105
	FileNameSuffixTimeFormat string
	//error level adapter will be saved in seperated error.adapter for keeping more time if it's ture
	SetErrFileHook bool
	//keep adapter count
	KeepCount int
	//adapter level
	LogLevel logrus.Level
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
		//keep adapter count
		KeepCount: 7,
		//adapter level
		LogLevel: logrus.DebugLevel,
	}
}

func (*OptLog) ConfigLogrus() (*logrus.Logger, error) {
	lg := logrus.New()
	return lg, nil
}
