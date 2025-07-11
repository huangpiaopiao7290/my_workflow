package logger

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/rs/zerolog"
)

var (
	loggerInstance zerolog.Logger
	lastFetchDay   int
	RequestDataKey = "requestData"
)

type GlobalRequestDataStruct struct {
	RequestID  string `json:"requestId"`  // 请求id
	RequestTime int64  `json:"requestTime"` // 请求时间
	ClientIP   string `json:"clientIp"`    // 客户端IP
	RemoteIP   string `json:"remoteIp"`    // 远程IP
	URL       string `json:"url"`         // 请求URL
	ContentType string `json:"contentType"` // 内容类型
	ContentSize int64  `json:"contentSize"` // 内容大小
	Topic      string `json:"topic"`      // 主题
}

// fieldSetter 定义设置日志字段的函数类型
type fieldSetter func(event *zerolog.Event, value interface{})

// fieldMappings 预定义结构体字段到日志设置器的映射
var fieldMappings = map[string]fieldSetter{
    "RequestID":     func(e *zerolog.Event, v interface{}) { e.Str("RequestID", v.(string)) },
    "RequestTime":   func(e *zerolog.Event, v interface{}) { e.Int64("RequestTime", v.(int64)) },
    "ClientIP":      func(e *zerolog.Event, v interface{}) { e.Str("ClientIP", v.(string)) },
    "RemoteIP":      func(e *zerolog.Event, v interface{}) { e.Str("RemoteIP", v.(string)) },
    "URL":           func(e *zerolog.Event, v interface{}) { e.Str("URL", v.(string)) },
    "ContentType":   func(e *zerolog.Event, v interface{}) { e.Str("ContentType", v.(string)) },
    "ContentSize":   func(e *zerolog.Event, v interface{}) { e.Int64("ContentSize", v.(int64)) },
    "Topic":         func(e *zerolog.Event, v interface{}) { e.Str("Topic", v.(string)) },
}


type requestDataHook struct{}

// 初始化全局日志记录
func initLogger() {
	today := time.Now().Format("20060102")
	fileName := "storage/loggs/app-" + today + ".log"
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Error init logger:", err)
		return
	}
	loggerInstance = zerolog.New(file).With().Logger().Hook(requestDataHook{})
}

// GetLogger 获取全局日志记录实例，每天只创建一个日志文件
func GetLogger() zerolog.Logger {
	current := time.Now().Day()
	if current != lastFetchDay {
		lastFetchDay = current
		initLogger()
	}
	return loggerInstance
}

// Run是requestDataHook实现方法，实现了zerolog.Hook接口
// 在日志事件发生的时候从事件的上下文中提取请求ID，添加到日志事件中
func (h requestDataHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
    if ctx := e.GetCtx(); ctx != nil {
        if requestData, ok := ctx.Value(RequestDataKey).(GlobalRequestDataStruct); ok {
            // 使用反射遍历结构体字段并应用预定义映射
            v := reflect.ValueOf(requestData)
            t := v.Type()
            
            for i := 0; i < v.NumField(); i++ {
                field := t.Field(i)
                fieldName := field.Name
                
                // 检查是否有预定义的映射
                if setter, exists := fieldMappings[fieldName]; exists {
                    // 获取字段值并传递给设置器
                    fieldValue := v.Field(i).Interface()
                    setter(e, fieldValue)
                }
            }
        }
    }
}
