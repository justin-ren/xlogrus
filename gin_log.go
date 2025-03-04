/*
 * @Author: justin-ren
 * @Date: 2025-03-01 02:50:20
 * @LastEditors: justin-ren
 * @LastEditTime: 2025-03-04 13:25:36
 * @FilePath: /xlogrus/gin_log.go
 * @Description: gin event log
 *
 */
package xlogrus

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	c "github.com/justin-ren/xlogrus/common"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type GinOpt struct {
	*c.OptLog
	SkipRoute map[string]struct{}
}

func GetGinOpt() *GinOpt {
	opt := GinOpt{OptLog: c.InitOpt()}
	opt.FileNamePrefix = "access.log"

	return &opt
}

func (opt *GinOpt) SetSkipRoute(r map[string]struct{}) error {
	opt.SkipRoute = r
	return nil
}

func WithSkipRoute[
	T any,
	PT interface {
		*T
		SetSkipRoute(map[string]struct{}) error
	},
](r map[string]struct{}) c.LogOption[T] {
	return c.NewLogOptionFunc(func(t *T) error {
		return PT(t).SetSkipRoute(r)
	})
}

func NewGinLog(setFunc ...c.LogOption[GinOpt]) (*TLogrus, gin.HandlerFunc, *GinOpt, error) {
	// 先初始化 GinOpt 结构体
	opt := GetGinOpt()

	//调用With函数修改GinOpt的默认值
	for _, f := range setFunc {
		if err := f.Apply(opt); err != nil {
			return nil, nil, nil, errors.Wrap(err, "failed to apply log option")
		}
	}

	if log, err := opt.ConfigLogrus(); err != nil {
		return log, nil, nil, errors.Cause(err)
	} else {
		return log,
			func(ctx *gin.Context) {
				if _, ok := opt.SkipRoute[ctx.Request.URL.Path]; ok {
					return
				}
				start := time.Now()
				path := ctx.Request.URL.Path
				raw := ctx.Request.URL.RawQuery
				ctx.Next()
				end := time.Now()
				latency := end.Sub(start) //记录请求处理时间
				clientIP := ctx.ClientIP()
				method := ctx.Request.Method
				statusCode := ctx.Writer.Status()
				//请求大小
				bodySize := ctx.Writer.Size()

				//记录url param
				if raw != "" {
					path = path + "?" + raw
				}
				//设置json字段内容
				entry := log.WithFields(logrus.Fields{
					"statusCode": statusCode,
					"latency":    latency, // time to process
					"clientIP":   clientIP,
					"method":     method,
					"path":       path,
					"dataLength": bodySize,
				})

				if len(ctx.Errors) > 0 {
					entry.Error(ctx.Errors.ByType(gin.ErrorTypePrivate).String())
				} else {
					//base on response value to match log level
					//msg := fmt.Sprintf("%s - \"%s %s\" %d %d (%dms)", clientIP, method, path, statusCode, bodySize, latency)
					if statusCode >= http.StatusInternalServerError { //500 assign to error level
						entry.Error()
					} else if statusCode >= http.StatusBadRequest { //400 assign to warn level
						entry.Warn()
					} else {
						entry.Info()
					}
				}
			},
			opt,
			nil
	}
}
