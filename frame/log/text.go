package log

import (
	"fmt"
	"strings"
	"time"
)

// TextFormatter 文本格式化器
type TextFormatter struct {
}

// Format 格式化日志
func (f *TextFormatter) Format(param *LoggingFormatParam) string {
	// 获取当前时间
	now := time.Now()
	// 格式化字段
	fieldsString := ""
	// 判断日志字段是否为空
	if param.LoggerFields != nil {
		//name=xx,age=xxx
		// 定义一个字符串构建器
		var sb strings.Builder
		var count = 0
		var lens = len(param.LoggerFields)
		// 遍历字段，将字段名和值拼接成字符串
		for k, v := range param.LoggerFields {
			// 拼接字段名和值
			fmt.Fprintf(&sb, "%s=%v", k, v)
			// 判断是否是最后一个字段，如果是则不添加逗号
			if count < lens-1 {
				fmt.Fprintf(&sb, ",")
				count++
			}
		}
		// 将字符串构建器转换为字符串
		fieldsString = sb.String()
	}
	var msgInfo = "\n msg: "
	// 判断日志级别，如果是错误级别，则添加错误信息
	if param.Level == LevelError {
		msgInfo = "\n Error Cause By: "
	}
	// 判断是否需要带颜色
	if param.IsColor {
		// 要带颜色  error的颜色 为红色 info为绿色 debug为蓝色
		levelColor := f.LevelColor(param.Level)
		// 为日志信息添加颜色
		msgColor := f.MsgColor(param.Level)
		return fmt.Sprintf("%s [frame] %s %s%v%s | level= %s %s %s%s%s %v %s %s ",
			yellow, reset, blue, now.Format("2006/01/02 - 15:04:05"), reset,
			levelColor, param.Level.Level(), reset, msgColor, msgInfo, param.Msg, reset, fieldsString,
		)
	}
	// 不带颜色直接返回
	return fmt.Sprintf("[frame] %v | level=%s%s%v %s",
		now.Format("2006/01/02 - 15:04:05"),
		param.Level.Level(), msgInfo, param.Msg, fieldsString)
}

// LevelColor 根据日志级别返回对应的颜色
func (f *TextFormatter) LevelColor(level LoggerLevel) string {
	// 根据日志级别返回对应的颜色
	switch level {
	case LevelDebug:
		return blue
	case LevelInfo:
		return green
	case LevelError:
		return red
	default:
		return cyan
	}
}

// MsgColor 根据日志信息级别返回对应的颜色
func (f *TextFormatter) MsgColor(level LoggerLevel) string {
	// 根据日志级别返回对应的颜色
	switch level {
	case LevelError:
		return red
	default:
		return ""
	}
}
