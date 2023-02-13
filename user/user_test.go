/**
 * @project xlogrus
 * @author justin-ren
 * @desc test user trace log
 * @date 10:35 PM 2/9/23
 **/

package user

import (
	"fmt"
	"github.com/itchyny/timefmt-go"               //convert golang time layout to linux time layout
	lTest "github.com/sirupsen/logrus/hooks/test" //logrus tools for test
	ast "github.com/stretchr/testify/assert"      //continue next case in case even failed
	req "github.com/stretchr/testify/require"     //exit if failed
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// check log msg from stdout and log file
func TestUserLog(t *testing.T) {
	//get init value
	opt := GetOpt()
	//create log file under /tmp/
	opt.LogPath = fmt.Sprintf("%v/logs/", os.TempDir())
	opt.FileNameSuffixTimeFormat = "%Y%m%d%H%M%S"
	//create logrus.Logger
	lg, err := New(opt)
	req.NoError(t, err)
	//creat hook for lastEntry in stdout
	hook := lTest.NewLocal(lg)
	//redirect screen output to /dev/null
	lg.SetOutput(io.Discard)
	//got full path for trace.log
	FileNamePrefix := fmt.Sprintf("%s%s", opt.LogPath, opt.FileNamePrefix)
	trcFile := fmt.Sprintf("%s.%s", FileNamePrefix, opt.FileNameSuffixTimeFormat)
	//got full path for error.log
	FileNamePrefix = fmt.Sprintf("%s%s", opt.LogPath, opt.ErrLogPrefix)
	errFile := fmt.Sprintf("%s.%s", FileNamePrefix, opt.ErrLogSuffix)
	//convert golang time layout to linux time layout which used in OptLog struct
	tm, err := timefmt.Parse(time.Now().Format("2006/01/02 15:04:05"), "%Y/%m/%d %H:%M:%S")
	req.NoError(t, err)

	//got trace.log and error.log with timestamp
	trcFile = timefmt.Format(tm, trcFile)
	errFile = timefmt.Format(tm, errFile)

	//create error log which write to stdout, trace.log and error.log
	lg.Error("Error Log")
	var lastEntry string
	//get log from stdout
	lastEntry, err = hook.LastEntry().String()
	ast.NoError(t, err)
	//should contain color in stdout
	req.Contains(t, lastEntry, "\x1b[0m \x1b[0;31mERROR\x1b[0m Error Log")

	//check log from trace.log and error.log
	for _, f := range [2]string{trcFile, errFile} {
		lContent, err := os.ReadFile(f)
		req.NoError(t, err)
		//get last line in log files
		arrLog := strings.Split(string(lContent), "\n")
		req.GreaterOrEqual(t, len(arrLog), 2)
		lastLine := arrLog[len(arrLog)-2]
		req.NoError(t, err)
		req.Contains(t, lastLine, "Error Log")
		req.NotContains(t, lastLine, "\u001B[0m \u001B[0;31mERROR\u001B[0m Error Log")
	}

}
