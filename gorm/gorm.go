/**
 * @project xlogrus
 * @author justin-ren
 * @desc create gorm middleware log which can be automatically called, log name is db.log by default
 * @date 3:30 PM 2/9/23
 **/

package gorm

import (
	"context"
	"fmt"
	c "github.com/justin-ren/xlogrus/common"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"strings"
	"time"
)

type OptGorm struct {
	SkipErrRecordNotFound bool
	SlowThreshold         time.Duration
	IsHelper              bool
	BKeywords             []BannedKeyword

	// if set to true, it will add latency information for your queries
	LogLatency bool

	//gorm log level for automatically triggering,
	//logrus log level is debug and should not be modified
	LogLevel logger.LogLevel

	//logrus parameters
	OptLogrus *c.OptLog
}

type LoggerGorm struct {
	Logger *logrus.Logger
	Opt    *OptGorm
}

// LogMode implementation log mode.
func (lm *LoggerGorm) LogMode(level logger.LogLevel) logger.Interface {
	lm.Opt.LogLevel = level
	return lm
}

func (lm *LoggerGorm) Info(ctx context.Context, msg string, args ...interface{}) {
	lm.Logger.WithContext(ctx).WithFields(logrus.Fields{"msg": lm.ignoreBKeyword(msg)}).Info(args...)
}

func (lm *LoggerGorm) Warn(ctx context.Context, msg string, args ...interface{}) {
	lm.Logger.WithContext(ctx).WithFields(
		logrus.Fields{
			"from": utils.FileWithLineNum(),
			"msg":  lm.ignoreBKeyword(msg),
		}).Warn(args...)
}

func (lm *LoggerGorm) Error(ctx context.Context, msg string, args ...interface{}) {
	lm.Logger.WithContext(ctx).WithFields(
		logrus.Fields{
			"from": utils.FileWithLineNum(),
			"msg":  lm.ignoreBKeyword(msg),
		}).Error(args...)
}

func (lm *LoggerGorm) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if lm.Opt.LogLevel <= logger.Silent {
		return
	}
	sql, rows := fc()

	fields := logrus.Fields{
		"from": utils.FileWithLineNum(),
		"rows": rows,
		"sql":  lm.ignoreBKeyword(sql),
	}

	elapsed := time.Since(begin)
	if lm.Opt.LogLatency {
		fields["elapsed"] = float64(elapsed.Nanoseconds()) / 1e6
	}
	switch {
	case err != nil && lm.Opt.LogLevel >= logger.Error &&
		(!errors.Is(err, gorm.ErrRecordNotFound) || lm.Opt.SkipErrRecordNotFound):

		fields["err"] = errors.Cause(err)
		lm.Logger.WithFields(fields).Error()

	case elapsed > lm.Opt.SlowThreshold && lm.Opt.SlowThreshold != 0 && lm.Opt.LogLevel >= logger.Warn:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", lm.Opt.SlowThreshold)
		fields["reason"] = slowLog
		lm.Logger.WithContext(ctx).WithFields(fields).Warn()

	case lm.Opt.LogLevel == logger.Info:
		lm.Logger.WithContext(ctx).WithFields(fields).Info()
	}
}

type BannedKeyword struct {
	// Keyword represent the string watched, for example : "password"
	Keyword string
	// CaseMatters if set to false, the Keyword matching will occur depending on the case.
	// if set to true, Keyword will cd .strictly match input messages
	IsCaseSensitive bool
}

func GetOpt() *OptGorm {
	opt := c.InitOpt()
	opt.FileNamePrefix = "db.log"
	return &OptGorm{
		true,
		500 * time.Millisecond,
		true,
		[]BannedKeyword{
			{
				"password",
				false,
			},
			{
				"pwd",
				false,
			},
		},
		true,
		logger.Warn,
		opt,
	}
}

func (lm *LoggerGorm) ignoreBKeyword(lContent string) string {
	if len(lm.Opt.BKeywords) <= 0 {
		return lContent
	}
	arrLine := strings.Split(strings.Trim(lContent, "\n"), "\n")
	for idx := 0; idx < len(lm.Opt.BKeywords); idx++ {
		for i := 0; i < len(arrLine); i++ {
			if lm.Opt.BKeywords[idx].IsCaseSensitive &&
				strings.Contains(arrLine[i], lm.Opt.BKeywords[idx].Keyword) {
				//found with case-sensitive
				arrLine[i] = fmt.Sprintf("ignored line with banned word %v",
					lm.Opt.BKeywords[idx].Keyword)
			} else if !lm.Opt.BKeywords[idx].IsCaseSensitive &&
				strings.Contains(
					strings.ToLower(arrLine[i]),
					strings.ToLower(lm.Opt.BKeywords[idx].Keyword),
				) { //found with ignore case-sensitive
				arrLine[i] = fmt.Sprintf("ignored line with banned word: %v",
					lm.Opt.BKeywords[idx].Keyword)
			}
		}
	}
	return strings.Join(arrLine, "\n")
}

func New(opt *OptGorm) (*LoggerGorm, error) {
	if lg, err := opt.OptLogrus.ConfigLogrus(); err != nil {
		return nil, errors.Cause(err)
	} else {
		return &LoggerGorm{
			lg,
			opt,
		}, nil
	}
}
