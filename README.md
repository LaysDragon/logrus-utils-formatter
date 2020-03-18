# logrus-utils-formatter
- github.com/go-errors/errors wrap stacktrace format 
  full stacktrace will only work under log.DebugLevel
- xerrors error trace format
- OpenTracing span Context LogFields
  - Error Report with StackTrace and detail from log.Error()
  - Caller info if `log.SetReportCaller(true)`


```go
package main
import (
    "github.com/LaysDragon/logrus-utils-formatter"
    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"
)

func init(){
	log.SetFormatter(&formatter.ErrorStracktraceTextFormatter{})
    log.SetReportCaller(true)
    log.SetLevel(log.InfoLevel)
}



func GinHandler(c *gin.Context) {
	log := log.WithContext(c) //c.Request.Context() contain OpenTracing span data
	if err:= DoSomeghing();err!= nil{
        log.WithError(err).WithError(err).Error("error happen while ...")
    }
    
}
```