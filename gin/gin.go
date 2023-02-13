/**
 * @project xlogrus
 * @author justin-ren
 * @desc create gin middleware log which can be automatically called,
 *       log name is access.log by default
 * @date 4:03 PM 2/9/23
 **/

package gin

import (
	"github.com/gin-gonic/gin"
	c "github.com/justin-ren/xlogrus/common"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type OptGin struct {
	//example 'user/logout', this route will be ignored in adapter
	SkipRoute map[string]struct{}
	OptLogrus *c.OptLog
}

func GetOpt() *OptGin {
	opt := c.InitOpt()
	opt.FileNamePrefix = "access.log"
	return &OptGin{
		OptLogrus: opt,
	}
}

func New(opt *OptGin) (*logrus.Logger, gin.HandlerFunc, error) {

	if log, err := opt.OptLogrus.ConfigLogrus(); err != nil {
		return log, func(ctx *gin.Context) {}, errors.Cause(err)
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
			nil
	}
}
