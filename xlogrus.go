/*
 * @Author: justin-ren
 * @Date: 2025-02-26 02:11:15
 * @LastEditors: justin-ren
 * @LastEditTime: 2025-03-03 23:22:06
 * @FilePath: /xlogrus-edit/xlogrus/xlogrus.go
 * @Description:
 *
 */
package xlogrus

import (
	c "github.com/justin-ren/xlogrus/common"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

type TLogrus = logrus.Logger
type TGormLog = logger.LogLevel

// type TGinHandleFunc = gin.HandlerFunc

// WithFileNamePrefix sets the filename prefix.
// 定义泛型函数 WithFileNamePrefix，用于创建设置文件名前缀的选项
func WithFileNamePrefix[
	// 定义两个泛型类型参数：
	T any, // T 可以是任意类型（比如 UserOpt）
	PT interface { // PT 必须满足以下两个条件：
		*T                              // 1. 必须是指向 T 的指针类型（比如 *UserOpt）
		SetFileNamePrefix(string) error // 2. 必须实现指定签名的方法
	},
](prefix string) c.LogOption[T] { // 接收字符串参数，返回 LogOption[T] 类型
	// 通过 NewLogOptionFunc 创建选项实例
	return c.NewLogOptionFunc(
		// 定义闭包函数作为选项的具体实现
		func(t *T) error { // 参数是 T 类型的指针（比如 *UserOpt）
			// 关键转换步骤：
			return PT(t).SetFileNamePrefix(prefix)
			// PT(t)：将 *T 转换为 PT 类型（指针类型转换，比如 *UserOpt -> *UserOpt）
			// 调用指针类型的 SetFileNamePrefix 方法
		},
	)
}

// WithLogLevel 设置日志级别
func WithLogLevel[
	T any,
	PT interface {
		*T
		SetLogLevel(string) error
	},
](logLevel string) c.LogOption[T] {
	return c.NewLogOptionFunc(func(t *T) error {
		return PT(t).SetLogLevel(logLevel)
	})
}

// WithStdoutTimeFormat 设置标准输出时间格式
func WithStdoutTimeFormat[
	T any,
	PT interface {
		*T
		SetStdoutTimeFormat(string) error
	},
](format string) c.LogOption[T] {
	return c.NewLogOptionFunc(func(t *T) error {
		return PT(t).SetStdoutTimeFormat(format)
	})
}

// WithLogFileTimeFormat 设置日志文件时间格式
func WithLogFileTimeFormat[
	T any,
	PT interface {
		*T
		SetLogFileTimeFormat(string) error
	},
](format string) c.LogOption[T] {
	return c.NewLogOptionFunc(func(t *T) error {
		return PT(t).SetLogFileTimeFormat(format)
	})
}

// WithLogPath 设置日志路径
func WithLogPath[
	T any,
	PT interface {
		*T
		SetLogPath(string) error
	},
](path string) c.LogOption[T] {
	return c.NewLogOptionFunc(func(t *T) error {
		return PT(t).SetLogPath(path)
	})
}

// WithFileNameSuffixTimeFormat 设置文件名后缀时间格式
func WithFileNameSuffixTimeFormat[
	T any,
	PT interface {
		*T
		SetFileNameSuffixTimeFormat(string) error
	},
](format string) c.LogOption[T] {
	return c.NewLogOptionFunc(func(t *T) error {
		return PT(t).SetFileNameSuffixTimeFormat(format)
	})
}

// WithSetErrFileHook 设置是否分离错误日志
func WithSetErrFileHook[
	T any,
	PT interface {
		*T
		SetSetErrFileHook(bool) error
	},
](enabled bool) c.LogOption[T] {
	return c.NewLogOptionFunc(func(t *T) error {
		return PT(t).SetSetErrFileHook(enabled)
	})
}

// WithKeepCount 设置日志保留数量
func WithKeepCount[
	T any,
	PT interface {
		*T
		SetKeepCount(int) error
	},
](count int) c.LogOption[T] {
	return c.NewLogOptionFunc(func(t *T) error {
		return PT(t).SetKeepCount(count)
	})
}

// WithErrLogPrefix 设置错误日志前缀
func WithErrLogPrefix[
	T any,
	PT interface {
		*T
		SetErrLogPrefix(string) error
	},
](prefix string) c.LogOption[T] {
	return c.NewLogOptionFunc(func(t *T) error {
		return PT(t).SetErrLogPrefix(prefix)
	})
}

// WithErrLogSuffix 设置错误日志后缀
func WithErrLogSuffix[
	T any,
	PT interface {
		*T
		SetErrLogSuffix(string) error
	},
](suffix string) c.LogOption[T] {
	return c.NewLogOptionFunc(func(t *T) error {
		return PT(t).SetErrLogSuffix(suffix)
	})
}
