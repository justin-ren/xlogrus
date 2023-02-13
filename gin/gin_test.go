/**
 * @project xlogrus
 * @author justin-ren
 * @desc test logrus for gin middleware with file access.log w/o color
 *      and stdout w/ color, skip route function
 * @date 7:32 PM 2/12/23
 **/

package gin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/itchyny/timefmt-go"               //convert golang time layout to linux time layout
	lTest "github.com/sirupsen/logrus/hooks/test" //logrus tools for test
	//ast "github.com/stretchr/testify/assert"      //continue next case in case even failed
	req "github.com/stretchr/testify/require" //exit if failed
)

type testRoute struct {
	name   string
	url    string
	want   string
	isWant bool
}

func ginLogHandle(t *testing.T, tr *testRoute, opt *OptGin, r *gin.Engine,
	hook *lTest.Hook) {
	t.Helper()
	r.GET(tr.url, func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"msg": tr.want})
	})
	//NewServer is really running a server and need to shut it down
	//httptest.NewServer(r)
	request := httptest.NewRequest("GET", tr.url, nil)
	response := httptest.NewRecorder()

	//To simulate one http server, don't need to shut it down
	r.ServeHTTP(response, request)

	//got response
	req.Equal(t, response.Code, http.StatusOK)
	req.Contains(t, response.Body.String(), tr.want)

	FileNamePrefix := fmt.Sprintf("%s%s", opt.OptLogrus.LogPath, opt.OptLogrus.FileNamePrefix)
	accessFile := fmt.Sprintf("%s.%s", FileNamePrefix, opt.OptLogrus.FileNameSuffixTimeFormat)

	lastEntry, err := hook.LastEntry().String()
	req.NoError(t, err)
	if tr.isWant {
		req.Contains(t, lastEntry, tr.url)
	} else {
		req.NotContains(t, lastEntry, tr.url)
	}

	//generate timestamp for access.log
	tm, err := timefmt.Parse(time.Now().Format("2006/01/02 15:04:05"), "%Y/%m/%d %H:%M:%S")
	req.NoError(t, err)

	//got trace.log and error.log with timestamp
	accessFile = timefmt.Format(tm, accessFile)
	lContent, err := os.ReadFile(accessFile)
	req.NoError(t, err)
	//get last line in log files
	if tr.isWant {
		req.Contains(t, string(lContent), tr.url)
	} else {
		req.NotContains(t, string(lContent), tr.url)
	}
}
func TestGinLog(t *testing.T) {

	tests := []testRoute{
		{"chkLogMsg",
			"/hello",
			"hello",
			true,
		},
		{
			"skipRoute",
			"/skip",
			"skip",
			false,
		},
	}
	var opt *OptGin
	var gLog *logrus.Logger
	var ginHandler gin.HandlerFunc
	var err error
	var r *gin.Engine
	var hook *lTest.Hook
	t.Run("initGinLog", func(t *testing.T) {
		opt = GetOpt()
		opt.OptLogrus.LogPath = fmt.Sprintf("%v/", os.TempDir())
		opt.SkipRoute = map[string]struct{}{
			"/skip": {},
		}
		gLog, ginHandler, err = New(opt)
		//redirect stdout to /dev/null
		gLog.SetOutput(io.Discard)
		//creat hook for lastEntry from buffer
		hook = lTest.NewLocal(gLog)
		req.NoError(t, err)
		r = gin.Default()
		gin.SetMode(gin.DebugMode)
		r.Use(ginHandler)
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ginLogHandle(t, &tt, opt, r, hook)
		})
	}
}
