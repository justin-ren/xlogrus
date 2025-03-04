/*
 * @Author: justin-ren
 * @Date: 2025-03-01 02:51:44
 * @LastEditors: justin-ren
 * @LastEditTime: 2025-03-04 10:11:43
 * @FilePath: /xlogrus/user_log.go
 * @Description: user debug log
 *
 */

package xlogrus

import (
	c "github.com/justin-ren/xlogrus/common"
	"github.com/pkg/errors"
)

type UserOpt struct {
	*c.OptLog //继承OptLog所有的方法
}

func GetUserOpt() *UserOpt {
	opt := UserOpt{c.InitOpt()}
	opt.FileNamePrefix = "trace.log"
	return &opt
}

func NewUserLog(setFunc ...c.LogOption[UserOpt]) (*TLogrus, *UserOpt, error) {
	opt := GetUserOpt()
	for _, f := range setFunc {
		if err := f.Apply(opt); err != nil {
			return nil, opt, errors.Wrap(err, "failed to apply log option")
		}
	}

	if lg, err := opt.ConfigLogrus(); err != nil {
		return lg, opt, errors.Cause(err)
	} else {
		return lg, opt, nil
	}
}
