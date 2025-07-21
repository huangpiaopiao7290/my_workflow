package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"my_workflow/pkg/common/constant"
	"os"
	"reflect"
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type contextKey string

var (
	loggerInstance zerolog.Logger
	lastFetchDay   int
	RequestDataKey contextKey = "requestData"
)

type GlobalRequestDataStruct struct {
	RequestID   string `json:"requestId"`   // 请求id
	RequestTime int64  `json:"requestTime"` // 请求时间
	ClientIP    string `json:"clientIp"`    // 客户端IP
	RemoteIP    string `json:"remoteIp"`    // 远程IP
	URL         string `json:"url"`         // 请求URL
	ContentType string `json:"contentType"` // 内容类型
	ContentSize int64  `json:"contentSize"` // 内容大小
	Topic       string `json:"topic"`       // 主题
}

// fieldSetter 定义设置日志字段的函数类型
type fieldSetter func(event *zerolog.Event, value any)

// fieldMappings 预定义结构体字段到日志设置器的映射
var fieldMappings = map[string]fieldSetter{
	"RequestID":   func(e *zerolog.Event, v any) { e.Str("RequestID", v.(string)) },
	"RequestTime": func(e *zerolog.Event, v any) { e.Int64("RequestTime", v.(int64)) },
	"ClientIP":    func(e *zerolog.Event, v any) { e.Str("ClientIP", v.(string)) },
	"RemoteIP":    func(e *zerolog.Event, v any) { e.Str("RemoteIP", v.(string)) },
	"URL":         func(e *zerolog.Event, v any) { e.Str("URL", v.(string)) },
	"ContentType": func(e *zerolog.Event, v any) { e.Str("ContentType", v.(string)) },
	"ContentSize": func(e *zerolog.Event, v any) { e.Int64("ContentSize", v.(int64)) },
	"Topic":       func(e *zerolog.Event, v any) { e.Str("Topic", v.(string)) },
}

type requestDataHook struct{}

// 初始化全局日志记录
func initLogger() {
	today := time.Now().Format("20060102")
	fileName := "storage/logs/app-" + today + ".log"
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
// @param:
// e *zerolog.Event: 日志事件对象，通过此对象可以访问和修改日志事件的数据。
// level zerolog.Level: 日志事件的级别，如Info、Error等。
// msg string: 日志事件的消息内容。
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

// 从上下文获取请求数据断言未为RequestDataStruct类型
// 参数：
//
//	context
//
// 返回值：
//
//	GlobalRequestDataStruct: 请求数据结构，包含特定的请求信息。
//	error: 如果提取或断言操作失败，则返回错误。
func GetRequestData(ctx context.Context) (GlobalRequestDataStruct, error) {
	val := ctx.Value(RequestDataKey)
	if val == nil {
		return GlobalRequestDataStruct{}, fmt.Errorf("request data not found in context")
	}

	if requestData, ok := val.(GlobalRequestDataStruct); ok {
		return requestData, nil
	}

	return GlobalRequestDataStruct{}, fmt.Errorf("invalid type for request data")
}

// GetRequestId 从上下文获取请求ID
// 如果无法获取请求数据，将会打印错误信息并生成一个新的UUID作为请求ID
// 参数:
//
//	ctx context.Context：上下文对象
//
// 返回值
//
//	string： 请求ID， 获取失败生成新的UUID
func GetRequestID(ctx context.Context) string {
	data, err := GetRequestData(ctx)
	if err != nil {
		fmt.Println("GetRequestID error", err.Error())
		return uuid.NewString()
	}
	return data.RequestID
}

func GetRequestTime(ctx context.Context) int64 {
	data, err := GetRequestData(ctx)
	if err != nil {
		fmt.Println("GetRequestTime error", err.Error())
		return time.Now().UnixMilli()
	}
	return data.RequestTime
}

// log 函数用于根据给定的日志级别和消息，以及可选的内容参数，记录结构化日志。
// 它通过获取全局日志实例，序列化内容参数，并根据日志级别选择合适的子日志记录器来实现日志记录。
// 此外，它还捕获调用者的程序计数器，方法名，文件名和行号，以便在日志中包含这些上下文信息。
func log(ctx context.Context, level zerolog.Level, msg string, args ...any) {
	logger := GetLogger() // 获取全局日志实例
	jsonContent, err := json.Marshal(args)
	if err != nil {
		return
	}
	var subLogger *zerolog.Event // 根据日志级别创建对应的日志记录器
	switch level {
	case zerolog.TraceLevel:
		subLogger = logger.Trace()
	case zerolog.DebugLevel:
		subLogger = logger.Debug()
	case zerolog.WarnLevel:
		subLogger = logger.Warn()
	case zerolog.ErrorLevel:
		subLogger = logger.Error()
	case zerolog.FatalLevel:
		subLogger = logger.Fatal()
	default:
		subLogger = logger.Info()
	}

	pc, _, _, _ := runtime.Caller(2)           // 获取调用者的信息
	methodName := runtime.FuncForPC(pc).Name() // 根据程序计数器获取方法名
	fileName, line := runtime.FuncForPC(pc).FileLine(pc)

	subLogger.Ctx(ctx).
		Str("Method", methodName).
		Str("File", fileName).
		Int("Line", line).
		Str("Message", msg).
		RawJSON("Content", jsonContent).
		Str("Timestamp", time.Now().Format("2006-01-02 15:04:05")).
		Msg("")
}

// Info 记录一条信息级别的日志
//
// 参数：
//   - message: 日志消息内容
//   - content: 可选参数，用于记录额外的信息，将按JSON格式序列化
func Info(ctx context.Context, message string, content ...any) {
	log(ctx, zerolog.InfoLevel, message, content...)
}

/*
Error 记录一条错误级别的日志

参数：
  - message: 日志消息内容
  - content: 可选参数，用于记录额外的信息，将按JSON格式序列化
*/
func Error(ctx context.Context, message string, content ...any) {
	log(ctx, zerolog.ErrorLevel, message, content...)
}

/*
Warn 记录一条警告级别的日志

参数：
  - message: 日志消息内容
  - content: 可选参数，用于记录额外的信息，将按JSON格式序列化
*/
func Warn(ctx context.Context, message string, content ...any) {
	log(ctx, zerolog.WarnLevel, message, content...)
}

func Fatal(ctx context.Context, message string, content ...any) {
	log(ctx, zerolog.FatalLevel, message, content...)
}

func Debug(ctx context.Context, message string, content ...interface{}) {
	env := viper.GetString("environment")
	if env == constant.EnvDev || env == constant.EnvTest {
		log(ctx, zerolog.DebugLevel, message, content...) // 调用Log函数记录信息级别的日志
	}
}
