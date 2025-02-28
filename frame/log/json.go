package log

import (
	"encoding/json"
	"fmt"
	"time"
)

// JsonFormatter json格式化输出
type JsonFormatter struct {
	TimeDisplay bool
}

func (f *JsonFormatter) Format(param *LoggingFormatParam) string {
	// 判断日志字段是否为空
	if param.LoggerFields == nil {
		// 创建一个空的字段
		param.LoggerFields = make(Fields)
	}

	now := time.Now()
	// 判断是否需要显示时间
	if f.TimeDisplay {
		param.LoggerFields["log_time"] = now.Format("2006/01/02 - 15:04:05")
	}
	// 添加日志级别和消息
	param.LoggerFields["msg"] = param.Msg
	// 添加日志级别
	param.LoggerFields["log_level"] = param.Level.Level()
	// 将字段转换为 JSON 字符串
	marshal, err := json.Marshal(param.LoggerFields)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s", string(marshal))
}
