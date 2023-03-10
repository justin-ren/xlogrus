/**
 * @project xlogrus
 * @author justin-ren
 * @desc create manual trace log by logrus with log name trace.log by default
 * @date 1:20 PM 2/9/23
 **/

package user

import (
	c "github.com/justin-ren/xlogrus/common"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func GetOpt() *c.OptLog {
	opt := c.InitOpt()
	opt.FileNamePrefix = "trace.log"
	return opt
}

func New(log *c.OptLog) (*logrus.Logger, error) {
	if lg, err := log.ConfigLogrus(); err != nil {
		return lg, errors.Cause(err)
	} else {
		return lg, nil
	}
}
