# xlogrus
Easy configuration with minor dependencies

## Features
### Support below logs
- manual user log with default name trace.log
- auto gin middleware log with default name access.log
- auto gorm middleware log with default name db.log

### Color enabled/disabled
- enabled color for screen
- disabled in log file

### Loop log with customized log count
- 7 files for trace.log/access.log/db.log by default

### Multi-hook for different log-level and middleware
- Centralized warn/error/fatal level to error.log 
- Seperated logs for user/gin/gorm to trace.log/access.log/db.log

### Alive logs with link
- link log point to alive log files, it's handy when using tail command

## Install
```bash
go get -u github.com/justin-ren/xlogrus
```


## Dependencies
```
//log engine
"github.com/sirupsen/logrus"
//loop log
"github.com/lestrrat-go/file-rotatelogs"
//hook for log file
"github.com/rifflock/lfshook"
//log format
"github.com/x-cray/logrus-prefixed-formatter"
```
## Example
### User Log
- example code
```golang
package main

import (
	"fmt"
	uLog "github.com/justin-ren/xlogrus/user"
	"github.com/pkg/errors"
)

func main() {
	opt := uLog.GetOpt()
	//path must end with '/', default path is './logs/'
	opt.LogPath = "/tmp/logs/"
	//below log full name is user.log.20230212, default is trace.log
	opt.FileNamePrefix = "user.log"
	if err := opt.SetXLevel("debug"); err != nil {
		fmt.Printf("%+v\n", errors.Cause(err))
		panic(err)
	}

	var lg *logrus.Logger
	var err error
	if lg, err = uLog.New(opt); err != nil {
		fmt.Printf("%+v\n", errors.Cause(err))
		panic(err)
	}
	//below log would be saved in user.log with timestamp %Y%m and keep 7 files by default
	//so automatically keep 7 days logs
	//To modify count of loop files by opt.KeepCount
	lg.Debugln("debug msg")
	lg.Infoln("info msg")
	//below log level would be saved in error.log with timestamp %Y%m and keep 7 files by default`
	//so automatically keep 7 months logs
	lg.Warnln("warn msg")
	lg.Errorln("error msg")
	lg.Fatalf("%+v", errors.New("error stack")) //save error stack to log files
}

```
- color is enabled in stdout
![Screenshot from 2023-02-14 08-12-05](https://user-images.githubusercontent.com/9739410/218624412-49ee8ab3-d418-44e1-9e03-7cf06918c835.png)
- color is disabled in log file
```bash
$ cat /tmp/logs/user.log
[2023-02-14 08:09:58.682854] DEBUG debug msg
[2023-02-14 08:09:58.683414]  INFO info msg
[2023-02-14 08:09:58.683469]  WARN warn msg
[2023-02-14 08:09:58.683670] ERROR error msg
[2023-02-14 08:09:58.683842] FATAL error stack
main.main
        /home/renxiong/projects/goTest/xlogrus-example/user.go:32
runtime.main
        /usr/local/go/src/runtime/proc.go:250
runtime.goexit
        /usr/local/go/src/runtime/asm_amd64.s:1598
$ cat /tmp/logs/error.log
[2023-02-13 23:35:15.254703] ERROR Error Log
[2023-02-14 08:09:58.683469]  WARN warn msg
[2023-02-14 08:09:58.683670] ERROR error msg
[2023-02-14 08:09:58.683842] FATAL error stack
main.main
        /home/renxiong/projects/goTest/xlogrus-example/user.go:32
runtime.main
        /usr/local/go/src/runtime/proc.go:250
runtime.goexit
        /usr/local/go/src/runtime/asm_amd64.s:1598
```


### Gin Log

- example code
```golang
func main() {

	opt := GinLog.GetOpt()
	//path must end with '/',default is ./logs/
	opt.LogPath = "/tmp/logs/"
	//full name of loop logs will be gin.log.202302013,default is access.log
	opt.FileNamePrefix = "gin.log"

	//timestamp of log file is defined as following
	//"%Y%m%d" is default
	//opt.FileNameSuffixTimeFormat = "%Y%m%d"
	//will not log info for route /skip
	opt.SkipRoute = map[string]struct{}{
		"/skip": {},
	}
	//keep cut of loop log is defined here
	//7 is default
	//opt.KeepCount = 7
	_, gLog, err := GinLog.New(opt)
	r := gin.New()
	r.Use(gLog, gin.Recovery())
	rLog := r.Group("log")
	rLog.GET("/hello", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"msg": "hello"})
	})
	rLog.GET("/skip", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"msg": "ignored"})
	})
	err = r.Run(":8080")
	if err != nil {
		panic(errors.Cause(err))
		return
	}
}
```
- color is enabled in stdout
![Screenshot from 2023-02-14 09-17-54](https://user-images.githubusercontent.com/9739410/218624450-03913136-7152-40c4-9e3b-2780d850eb9a.png)
- color is disabled in log file
```bash
$ curl localhost:8080/hello
404 page not found$ curl localhost:8080/log/hello
{"msg":"hello"}$ curl localhost:8080/log/skip
{"msg":"ignored"}$ ls -lrt /tmp/logs/
total 12
-rw-r--r-- 1 renxiong renxiong 405 Feb 14 08:09 user.log.20230214
lrwxrwxrwx 1 renxiong renxiong  17 Feb 14 08:09 user.log -> user.log.20230214
lrwxrwxrwx 1 renxiong renxiong  16 Feb 14 09:13 gin.log -> gin.log.20230214
-rw-r--r-- 1 renxiong renxiong 482 Feb 14 09:13 error.log.202302
lrwxrwxrwx 1 renxiong renxiong  16 Feb 14 09:13 error.log -> error.log.202302
-rw-r--r-- 1 renxiong renxiong 379 Feb 14 09:15 gin.log.20230214
$ cat /tmp/logs/gin.log
[2023-02-14 09:13:21.370334]  WARN  clientIP=127.0.0.1 dataLength=-1 latency=264ns method=GET path=/hello statusCode=404
[2023-02-14 09:14:59.056411]  INFO  clientIP=127.0.0.1 dataLength=15 latency=1.070547ms method=GET path=/log/hello statusCode=200
[2023-02-14 09:15:26.864735]  INFO  clientIP=127.0.0.1 dataLength=17 latency=32.015µs method=GET path=/log/skip statusCode=200
$ cat /tmp/logs/error.log | grep clientIP
[2023-02-14 09:13:21.370334]  WARN  clientIP=127.0.0.1 dataLength=-1 latency=264ns method=GET path=/hello statusCode=404
$ cat /tmp/logs/error.log | tail -1
[2023-02-14 09:13:21.370334]  WARN  clientIP=127.0.0.1 dataLength=-1 latency=264ns method=GET path=/hello statusCode=404
$ 
```

### Gorm log
- example code
```golang
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"

	// GinLog gin middleware auto log
	GinLog "github.com/justin-ren/xlogrus/gin"
)

func main() {

	opt := GinLog.GetOpt()
	//path must end with '/',default is ./logs/
	opt.LogPath = "/tmp/logs/"
	//full name of loop logs will be gin.log.202302013,default is access.log
	opt.FileNamePrefix = "gin.log"

	//timestamp of log file is defined as following
	//"%Y%m%d" is default
	opt.FileNameSuffixTimeFormat = "%Y%m%d"
	//will not log info for route /skip
	opt.SkipRoute = map[string]struct{}{
		"/skip": {},
	}
	//keep cut of loop log is defined here
	//7 is default
	opt.KeepCount = 7
	_, gLog, err := GinLog.New(opt)
	r := gin.New()
	r.Use(gLog, gin.Recovery())
	rLog := r.Group("log")
	rLog.GET("/hello", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"msg": "hello"})
	})
	rLog.GET("/skip", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"msg": "ignored"})
	})
	err = r.Run(":8080")
	if err != nil {
		panic(errors.Cause(err))
		return
	}
}
```



- color is enabled in stdout
![Screenshot from 2023-02-14 09-46-15](https://user-images.githubusercontent.com/9739410/218624507-f94362ce-26d1-40d3-937a-eb1036ea2c1b.png)
- color is disabled in log file

```bash
$ cat /tmp/logs/gorm.log
[2023-02-14 09:45:43.169051] ERROR  elapsed=0.150981 err=no such table: not_existing_tables from=/home/renxiong/projects/goTest/xlogrus-example/gorm.go:56 rows=0 sql=INSERT INTO `not_existing_tables` DEFAULT VALUES
$ cat /tmp/logs/error.log | tail -1
[2023-02-14 09:45:43.169051] ERROR  elapsed=0.150981 err=no such table: not_existing_tables from=/home/renxiong/projects/goTest/xlogrus-example/gorm.go:56 rows=0 sql=INSERT INTO `not_existing_tables` DEFAULT VALUES
```
