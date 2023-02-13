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

### Multi-hook for different log-level/middleware
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


