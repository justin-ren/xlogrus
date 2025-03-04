/**
 * @project xlogrus
 * @author justin-ren
 * @desc test banned keyword for sensitive info such as password
 *       stdout log w/ color, file log w/o color
 * @date 2:45 PM 2/13/23
 **/

package xlogrus

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/itchyny/timefmt-go"               //convert time layout from linux to golang
	lTest "github.com/sirupsen/logrus/hooks/test" //logrus tools for test
	ast "github.com/stretchr/testify/assert"      //continue next code in case even failed
	req "github.com/stretchr/testify/require"     //exit if failed
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func initGormLog(t *testing.T) *GormLog {
	t.Helper()
	//  opt := GetGormOpt()
	BKeywords := []BannedKeyword{
		{
			"pwd", true,
		},
		{
			"password", false,
		},
	}
	path := fmt.Sprintf("%v/logs/", os.TempDir())
	lg, _, err := NewGormLog(
		WithBKeywords[GormOpt](BKeywords),
		WithLogPath[GormOpt](path),
	)
	req.NoError(t, err)
	return lg
}
func TestIgnoreBKeyword(t *testing.T) {
	var lg *GormLog
	t.Run("InitLogGorm", func(t *testing.T) {
		lg = initGormLog(t)
	})
	type args struct {
		lContent string
	}
	tests := []struct {
		name   string
		args   args
		want   string
		isWant bool
	}{
		{"Ignored",
			args{
				`wrong Password 123
 failed with pwd 456`,
			},
			"ignored line",
			true,
		},
		{"NotIgnored",
			args{
				`wrong Pass 123
 failed with Pwd 456`,
			},
			"ignored line",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lg.ignoreBKeyword(tt.args.lContent)
			if tt.isWant {
				ast.Equal(t, strings.Contains(got, tt.want), true)
			} else {
				ast.NotEqual(t, strings.Contains(got, tt.want), true)
			}
		})
	}
}

func TestGormLog(t *testing.T) {
	lm := initGormLog(t)
	hook := lTest.NewLocal(lm.Logger)
	//redirect screen output to /dev/null
	lm.Logger.SetOutput(io.Discard)
	//got full path for db.log
	FileNamePrefix := fmt.Sprintf("%s%s", lm.Opt.LogPath, lm.Opt.FileNamePrefix)
	dbFile := fmt.Sprintf("%s.%s", FileNamePrefix, lm.Opt.FileNameSuffixTimeFormat)
	//got full path for error.log
	// FileNamePrefix = fmt.Sprintf("%s%s", lm.Opt.OptLogrus.LogPath, lm.Opt.OptLogrus.ErrLogPrefix)
	// errFile := fmt.Sprintf("%s.%s", FileNamePrefix, lm.Opt.OptLogrus.ErrLogSuffix)
	//convert golang time layout to linux time layout which used in OptLog struct
	tm, err := timefmt.Parse(time.Now().Format("2006/01/02 15:04:05"), "%Y/%m/%d %H:%M:%S")
	req.NoError(t, err)

	//got db.log and error.log with timestamp
	dbFile = timefmt.Format(tm, dbFile)
	// errFile = timefmt.Format(tm, errFile)

	db, err := gorm.Open(sqlite.Open(
		genSqliteConn(t)),
		&gorm.Config{Logger: lm},
	)
	// check if database correctly created
	req.NoError(t, err)
	req.NotNil(t, db)

	sqlDB, err := db.DB()
	req.NoError(t, err)
	req.NotNil(t, sqlDB)

	defer func() {
		ast.NoError(t, sqlDB.Close())
	}()

	// NotExistingTable is a simple empty struct that does not exist in current database,
	// so if we try to create a new entry of this struct, gorm must return an error
	// telling us that this table does not exist
	type NotExistingTable struct{}

	errCreate := db.Create(&NotExistingTable{}).Error
	//t.Log(errCreate.Error())

	// testing gorm is not a purpose of this test, but to ensure consistency we
	// must check if errCreate is not empty
	req.NotEmpty(t, errCreate)
	req.Contains(t, errCreate.Error(), "no such table")
	req.Contains(t, errCreate.Error(), "not_existing_tables")

	ast.Equal(t, 1, len(hook.Entries))
	req.NotNil(t, hook.LastEntry())

	lastEntry, err := hook.LastEntry().String()
	req.NoError(t, err)
	lContent, err := os.ReadFile(dbFile)
	req.NoError(t, err)
	dbContent := string(lContent)

	// lContent, err = os.ReadFile(dbFile)
	req.NoError(t, err)
	errContent := string(lContent)
	tests := []struct {
		name   string
		target string
		want   string
	}{
		{
			"stdoutContent",
			lastEntry,
			errCreate.Error(),
		},
		{
			"stdoutColor",
			lastEntry,
			"\x1b[0m \x1b[0;31mERROR\x1b[0m",
		},
		{
			"dbLogFile",
			dbContent,
			"ERROR",
		},
		{
			"errLogFile",
			errContent,
			"ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//fmt.Println(tt.target)
			ast.Contains(t, tt.target, tt.want)
		})
	}

}

func genSqliteConn(t *testing.T) string {
	t.Helper()
	const sqliteConnString = "file:%s?mode=memory&cache=shared"

	return fmt.Sprintf(sqliteConnString, t.Name())
}
