/*
 * @Author: justin-ren
 * @Date: 2025-03-01 02:52:00
 * @LastEditors: justin-ren
 * @LastEditTime: 2025-03-04 13:31:12
 * @FilePath: /xlogrus/gorm_log.go
 * @Description: gorm event log
 *
 */
package xlogrus

import (
	"context"
	"fmt"
	"strings"
	"time"

	c "github.com/justin-ren/xlogrus/common"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type GormLog struct {
	Logger *logrus.Logger
	Opt    *GormOpt
}

// LogMode implementation log mode.
func (gormLog *GormLog) LogMode(level logger.LogLevel) logger.Interface {
	gormLog.Opt.GormLogLevel = level
	return gormLog
}

func (gormLog *GormLog) Info(ctx context.Context, msg string, args ...interface{}) {
	gormLog.Logger.WithContext(ctx).WithFields(logrus.Fields{"msg": gormLog.ignoreBKeyword(msg)}).Info(args...)
}

func (gormLog *GormLog) Warn(ctx context.Context, msg string, args ...interface{}) {
	gormLog.Logger.WithContext(ctx).WithFields(
		logrus.Fields{
			"from": utils.FileWithLineNum(),
			"msg":  gormLog.ignoreBKeyword(msg),
		}).Warn(args...)
}

func (gormLog *GormLog) Error(ctx context.Context, msg string, args ...interface{}) {
	gormLog.Logger.WithContext(ctx).WithFields(
		logrus.Fields{
			"from": utils.FileWithLineNum(),
			"msg":  gormLog.ignoreBKeyword(msg),
		}).Error(args...)
}

func (gormLog *GormLog) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if gormLog.Opt.GormLogLevel <= logger.Silent {
		return
	}
	sql, rows := fc()

	fields := logrus.Fields{
		"from": utils.FileWithLineNum(),
		"rows": rows,
		"sql":  gormLog.ignoreBKeyword(sql),
	}

	elapsed := time.Since(begin)
	if gormLog.Opt.LogLatency {
		fields["elapsed"] = float64(elapsed.Nanoseconds()) / 1e6
	}
	switch {
	case err != nil && gormLog.Opt.GormLogLevel >= logger.Error &&
		(!errors.Is(err, gorm.ErrRecordNotFound) || gormLog.Opt.SkipErrRecordNotFound):

		fields["err"] = errors.Cause(err)
		gormLog.Logger.WithFields(fields).Error()

	case elapsed > gormLog.Opt.SlowThreshold && gormLog.Opt.SlowThreshold != 0 && gormLog.Opt.GormLogLevel >= logger.Warn:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", gormLog.Opt.SlowThreshold)
		fields["reason"] = slowLog
		gormLog.Logger.WithContext(ctx).WithFields(fields).Warn()

	case gormLog.Opt.GormLogLevel == logger.Info:
		gormLog.Logger.WithContext(ctx).WithFields(fields).Info()
	}
}

type BannedKeyword struct {
	// Keyword represent the string watched, for example : "password"
	Keyword string
	// CaseMatters if set to false, the Keyword matching will occur depending on the case.
	// if set to true, Keyword will cd .strictly match input messages
	IsCaseSensitive bool
}

/*GetOpt
 * @msg get default values
 * @return: *GormOpt
 */
func GetGormOpt() *GormOpt {
	opt := c.InitOpt()
	opt.FileNamePrefix = "db.log"
	return &GormOpt{
		SkipErrRecordNotFound: true,
		SlowThreshold:         500 * time.Millisecond,
		IsHelper:              true,
		BKeywords: []BannedKeyword{
			{
				"password",
				false,
			},
			{
				"pwd",
				false,
			},
		},
		LogLatency:   true,
		GormLogLevel: logger.Warn,
		OptLog:       opt,
	}
}

/*ignoreBKeyword
 * @msg deal with sensitive word, and replaced that line with "ignore line with banned word..."
 * @receiver gormLog
 * @param lContent
 * @return: string
 */
func (gormLog *GormLog) ignoreBKeyword(lContent string) string {
	if len(gormLog.Opt.BKeywords) <= 0 {
		return lContent
	}
	arrLine := strings.Split(strings.Trim(lContent, "\n"), "\n")
	for idx := 0; idx < len(gormLog.Opt.BKeywords); idx++ {
		for i := 0; i < len(arrLine); i++ {
			if gormLog.Opt.BKeywords[idx].IsCaseSensitive &&
				strings.Contains(arrLine[i], gormLog.Opt.BKeywords[idx].Keyword) {
				//found with case-sensitive
				arrLine[i] = fmt.Sprintf("ignored line with banned word %v",
					gormLog.Opt.BKeywords[idx].Keyword)
			} else if !gormLog.Opt.BKeywords[idx].IsCaseSensitive &&
				strings.Contains(
					strings.ToLower(arrLine[i]),
					strings.ToLower(gormLog.Opt.BKeywords[idx].Keyword),
				) { //found with ignore case-sensitive
				arrLine[i] = fmt.Sprintf("ignored line with banned word: %v",
					gormLog.Opt.BKeywords[idx].Keyword)
			}
		}
	}
	return strings.Join(arrLine, "\n")
}

type GormOpt struct {
	//ignore if NotFound error happened
	SkipErrRecordNotFound bool
	//slow sql threshold
	SlowThreshold time.Duration
	//record line number and filename
	IsHelper bool
	//replace sensitive word, such as password
	BKeywords []BannedKeyword

	// if set to true, it will add latency information for your queries
	LogLatency bool
	//gorm log level for automatically triggering
	//logrus log level is debug and don't need to modify
	GormLogLevel logger.LogLevel
	*c.OptLog
}

// SetGormLoglevel 方法
// level支持 silent, error, warn, warning,info
func (g *GormOpt) SetGormLoglevel(level string) error {
	switch level {
	case "silent":
		g.GormLogLevel = logger.Silent
	case "error":
		g.GormLogLevel = logger.Error
	case "warn", "warning": // 支持 "warn" 和 "warning"
		g.GormLogLevel = logger.Warn
	case "info":
		g.GormLogLevel = logger.Info
	default:
		return errors.New("invalid log level")
	}
	return nil
}

// SetSkipErrRecordNotFound 设置 SkipErrRecordNotFound 字段
func (g *GormOpt) SetSkipErrRecordNotFound(value bool) error {
	g.SkipErrRecordNotFound = value
	return nil
}

// SetSlowThreshold 设置 SlowThreshold 字段
func (g *GormOpt) SetSlowThreshold(threshold time.Duration) error {
	if threshold < 0 {
		return errors.New("slow threshold cannot be negative")
	}
	g.SlowThreshold = threshold
	return nil
}

// SetIsHelper 设置 IsHelper 字段
func (g *GormOpt) SetIsHelper(value bool) error {
	g.IsHelper = value
	return nil
}

// SetBKeywords 设置 BKeywords 字段
func (g *GormOpt) SetBKeywords(keywords []BannedKeyword) error {
	if len(keywords) == 0 {
		return errors.New("banned keywords list cannot be empty")
	}
	g.BKeywords = keywords
	return nil
}

// SetLogLatency 设置 LogLatency 字段
func (g *GormOpt) SetLogLatency(value bool) error {
	g.LogLatency = value
	return nil
}

// WithGormLogLevel 设置 GORM 日志级别
func WithGormLogLevel[
	T any,
	PT interface {
		*T
		SetGormLoglevel(string) error
	},
](level string) c.LogOption[T] {
	return c.NewLogOptionFunc(func(t *T) error {
		return PT(t).SetGormLoglevel(level)
	})
}

// 为 GormOpt 的字段生成 With 函数
// WithSkipErrRecordNotFound 设置是否忽略记录未找到错误
func WithSkipErrRecordNotFound[
	T any,
	PT interface {
		*T
		SetSkipErrRecordNotFound(bool) error
	},
](value bool) c.LogOption[T] {
	return c.NewLogOptionFunc(func(t *T) error {
		return PT(t).SetSkipErrRecordNotFound(value)
	})
}

// WithSlowThreshold 设置慢查询阈值
func WithSlowThreshold[
	T any,
	PT interface {
		*T
		SetSlowThreshold(time.Duration) error
	},
](threshold time.Duration) c.LogOption[T] {
	return c.NewLogOptionFunc(func(t *T) error {
		return PT(t).SetSlowThreshold(threshold)
	})
}

// WithIsHelper 设置是否为帮助函数
func WithIsHelper[
	T any,
	PT interface {
		*T
		SetIsHelper(bool) error
	},
](value bool) c.LogOption[T] {
	return c.NewLogOptionFunc(func(t *T) error {
		return PT(t).SetIsHelper(value)
	})
}

// WithBKeywords 设置禁用关键词列表
func WithBKeywords[
	T any,
	PT interface {
		*T
		SetBKeywords([]BannedKeyword) error
	},
](keywords []BannedKeyword) c.LogOption[T] {
	return c.NewLogOptionFunc(func(t *T) error {
		return PT(t).SetBKeywords(keywords)
	})
}

// WithLogLatency 设置是否记录延迟
func WithLogLatency[
	T any,
	PT interface {
		*T
		SetLogLatency(bool) error
	},
](value bool) c.LogOption[T] {
	return c.NewLogOptionFunc(func(t *T) error {
		return PT(t).SetLogLatency(value)
	})
}
func NewGormLog(setFunc ...c.LogOption[GormOpt]) (*GormLog, *GormOpt, error) {
	opt := GetGormOpt()

	//调用With函数修改GormOpt的默认值
	for _, f := range setFunc {
		if err := f.Apply(opt); err != nil {
			return nil, nil, errors.Wrap(err, "failed to apply log option")
		}
	}

	if lg, err := opt.ConfigLogrus(); err != nil {
		return nil, nil, errors.Cause(err)
	} else {
		return &GormLog{
			lg,
			opt,
		}, opt, nil
	}
}
