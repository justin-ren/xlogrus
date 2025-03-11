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


func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "user" {
			userLogTest()
		} else if os.Args[1] == "gin" {
			ginLogTest()
		} else if os.Args[1] == "gorm" {
			gormLogTest()
		}
	} else {
		userLogTest()
	}
}

func userLogTest() {
	var lg *xlog.TLogrus
	var err error
	if lg, _, err = xlog.NewUserLog(
		xlog.WithFileNamePrefix[xlog.UserOpt]("user.log"), // 自动继承方法
		xlog.WithLogLevel[xlog.UserOpt]("info"),
		xlog.WithLogPath[xlog.UserOpt]("/tmp/logs/"),
	); err != nil {
		fmt.Printf("%+v\n", errors.Cause(err))
		panic(err)
	}

	lg.Debugln("debug msg")
	lg.Infoln("info msg")
	//below log level would be saved in error.log with timestamp %Y%m and keep 7 files by default`
	//so automatically keep 7 months logs
	lg.Warnln("warn msg")
	lg.Errorln("error msg")
	lg.Fatalf("%+v", errors.New("error stack")) //save error stack to log filess
}

```
- color is enabled in stdout
![user log](https://github.com/user-attachments/assets/a20a21cf-80c0-483c-b0cd-76ecf25c9bb8)
- color is disabled in log file
```bash
vscode ➜ /workspaces/go/xlogrus-edit (master) $ cat /tmp/logs/user.log
[2025-03-04 08:25:14.153871]  INFO info msg
[2025-03-04 08:25:14.154505]  WARN warn msg
[2025-03-04 08:25:14.154887] ERROR error msg
[2025-03-04 08:25:14.155100] FATAL error stack
main.userLogTest
        /workspaces/go/xlogrus-edit/main.go:57
main.main
        /workspaces/go/xlogrus-edit/main.go:28
runtime.main
        /usr/local/go/src/runtime/proc.go:271
runtime.goexit
        /usr/local/go/src/runtime/asm_amd64.s:1695
```


### Gin Log

- example code
```golang
//visit http://localhost:8080/log/skip and http://localhost:8080/log/hello for test
func ginLogTest() {
	var gHandle gin.HandlerFunc
	var err error
	if _, gHandle, _, err = xlog.NewGinLog(
		xlog.WithFileNamePrefix[xlog.GinOpt]("access.log"), // 自动继承方法
		xlog.WithLogLevel[xlog.GinOpt]("info"),
		xlog.WithLogPath[xlog.GinOpt]("/tmp/logs/"),
		xlog.WithSkipRoute[xlog.GinOpt](map[string]struct{}{"/skip": {}}),
	); err != nil {
		fmt.Printf("%+v\n", errors.Cause(err))
		panic(err)
	}
	r := gin.New()
	r.Use(gHandle, gin.Recovery())
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
	}

}

```
- color is enabled in stdout
![gin log](https://github.com/user-attachments/assets/d6620b78-cafc-4a57-a805-3f8847c96384)
- color is disabled in log file
```bash
vscode ➜ /workspaces/go/xlogrus-edit (master) $ cat /tmp/logs/error.log
[2025-03-03 23:38:20.619887]  WARN  clientIP=127.0.0.1 dataLength=-1 latency=7.5µs method=GET path=/ statusCode=404
vscode ➜ /workspaces/go/xlogrus-edit (master) $ cat /tmp/logs/access.log
[2025-03-03 23:38:20.619887]  WARN  clientIP=127.0.0.1 dataLength=-1 latency=7.5µs method=GET path=/ statusCode=404
[2025-03-03 23:38:25.912537]  INFO  clientIP=127.0.0.1 dataLength=17 latency=45.2µs method=GET path=/log/skip statusCode=200
[2025-03-03 23:38:42.127364]  INFO  clientIP=127.0.0.1 dataLength=15 latency=20µs method=GET path=/log/hello statusCode=200
```

### Gorm log
- example code
```golang
func gormLogTest() {
	type notExistingTable struct{}
	connString := fmt.Sprintf("file:%s?mode=memory&cache=shared", "gormLogTest")
	var lg *xlog.GormLog
	var err error
	if lg, _, err = xlog.NewGormLog(
		xlog.WithErrLogPrefix[xlog.GormOpt]("db.log"),
		xlog.WithBKeywords[xlog.GormOpt]([]xlog.BannedKeyword{
			{
				Keyword:         "pass",
				IsCaseSensitive: false,
			},
		}),
		xlog.WithGormLogLevel[xlog.GormOpt]("warn"),
		xlog.WithLogPath[xlog.GormOpt]("/tmp/logs/"),
	); err != nil {
		fmt.Printf("%+v\n", errors.Cause(err))
		panic(err)
	}

	var db *gorm.DB
	if db, err = gorm.Open(sqlite.Open(
		connString),
		&gorm.Config{Logger: lg},
	); err != nil {
		fmt.Printf("%+v\n", errors.Cause(err))
		panic(err)
	}
	var sqlDB *sql.DB
	if sqlDB, err = db.DB(); err != nil {
		fmt.Printf("%+v\n", errors.Cause(err))
		panic(err)
	}

	defer func() {
		if err := sqlDB.Close(); err != nil {
			fmt.Printf("%+v\n", errors.Cause(err))
			panic(err)
		}
	}()

	if errCreate := db.Create(&notExistingTable{}).Error; errCreate != nil {
		fmt.Printf("failed to create table as expected: %+v\n", errCreate)
	}

}
```



- color is enabled in stdout
![gorm](https://github.com/user-attachments/assets/bc7ecc93-f1d7-427e-8d88-2d5c474cbb5f)
- color is disabled in log file

```bash
vscode ➜ /workspaces/go/xlogrus-edit (master) $ cat /tmp/logs/db.log
[2025-03-04 08:16:22.290016] ERROR  elapsed=0.204776 err=no such table: not_existing_tables from=/workspaces/go/xlogrus-edit/main.go:129 rows=0 sql=INSERT INTO `not_existing_tables` DEFAULT VALUES
```
