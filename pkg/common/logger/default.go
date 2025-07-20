package logger

import (
	"context"
	"time"

	"github.com/google/uuid"
)


func InitLogCtx(ctx context.Context) context.Context {
	data := GlobalRequestDataStruct{
		RequestID: uuid.NewString(),
		RequestTime: time.Now().UnixNano(),
	}
	return context.WithValue(ctx, RequestDataKey, data)
}