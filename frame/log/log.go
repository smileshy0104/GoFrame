package log

import (
	"fmt"
	"frame/internal/string_func"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

// 颜色代码，用于控制台输出的着色
const (
	greenBg   = "\033[97;42m"
	whiteBg   = "\033[90;47m"
	yellowBg  = "\033[90;43m"
	redBg     = "\033[97;41m"
	blueBg    = "\033[97;44m"
	magentaBg = "\033[97;45m"
	cyanBg    = "\033[97;46m"
	green     = "\033[32m"
	white     = "\033[37m"
	yellow    = "\033[33m"
	red       = "\033[31m"
	blue      = "\033[34m"
	magenta   = "\033[35m"
	cyan      = "\033[36m"
	reset     = "\033[0m"
)

// LoggerLevel 定义日志级别类型
type LoggerLevel int

// Level 返回日志级别的字符串表示形式
func (l LoggerLevel) Level() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	default:
		return ""
	}
}

// 日志级别常量
const (
	// 使用 LoggerLevel 定义日志级别
	LevelDebug LoggerLevel = iota
	LevelInfo
	LevelError
)

// Fields 是一个键值对集合，用于存储日志字段
type Fields map[string]any

// Logger 是日志记录器结构体
type Logger struct {
	Formatter    LoggingFormatter // 日志格式化接口
	Level        LoggerLevel      // 日志级别
	Outs         []*LoggerWriter  // 输出目标列表
	LoggerFields Fields           // 日志字段
	logPath      string           // 日志文件路径
	LogFileSize  int64            // 单个日志文件的最大大小
}

// LoggerWriter 表示日志输出目标
type LoggerWriter struct {
	Level LoggerLevel // 输出的日志级别
	Out   io.Writer   // 输出流
}

// LoggingFormatter 是日志格式化的接口
type LoggingFormatter interface {
	Format(param *LoggingFormatParam) string
}

// LoggingFormatParam 是日志格式化的参数结构体
type LoggingFormatParam struct {
	Level        LoggerLevel // 日志级别
	IsColor      bool        // 是否使用颜色
	LoggerFields Fields      // 日志字段
	Msg          any         // 日志消息
}

// LoggerFormatter 实现了 LoggingFormatter 接口
type LoggerFormatter struct {
	Level        LoggerLevel // 日志级别
	IsColor      bool        // 是否使用颜色
	LoggerFields Fields      // 日志字段
}

// New 创建一个新的 Logger 实例
func New() *Logger {
	return &Logger{}
}

// Default 创建并返回一个带有默认配置的 Logger 实例
func Default() *Logger {
	logger := New()
	logger.Level = LevelDebug
	w := &LoggerWriter{
		Level: LevelDebug,
		Out:   os.Stdout,
	}
	logger.Outs = append(logger.Outs, w)
	logger.Formatter = &TextFormatter{} // 假设 TextFormatter 是一个实现了 LoggingFormatter 的结构体
	return logger
}

// Info 记录一条 INFO 级别的日志
func (l *Logger) Info(msg any) {
	// 调用 Print 方法记录一条 INFO 级别的日志
	l.Print(LevelInfo, msg)
}

// Debug 记录一条 DEBUG 级别的日志
func (l *Logger) Debug(msg any) {
	// 调用 Print 方法记录一条 DEBUG 级别的日志
	l.Print(LevelDebug, msg)
}

// Error 记录一条 ERROR 级别的日志
func (l *Logger) Error(msg any) {
	// 调用 Print 方法记录一条 ERROR 级别的日志
	l.Print(LevelError, msg)
}

// Print 根据指定的日志级别和消息打印日志
func (l *Logger) Print(level LoggerLevel, msg any) {
	// 检查当前日志级别是否高于输入级别
	if l.Level > level {
		// 如果当前日志级别高于输入级别，则不打印日志
		return
	}
	// 创建 LoggingFormatParam
	param := &LoggingFormatParam{
		Level:        level,
		LoggerFields: l.LoggerFields,
		Msg:          msg,
	}
	// 使用 Formatter 格式化日志信息
	str := l.Formatter.Format(param)
	// 遍历所有输出目的地
	for _, out := range l.Outs {
		// 如果输出目的地是标准输出，则启用颜色模式
		if out.Out == os.Stdout {
			param.IsColor = true
			// 重新格式化日志信息以应用颜色模式
			str = l.Formatter.Format(param)
			fmt.Fprintln(out.Out, str)
		}
		// 如果输出级别的设置为 -1 或与当前日志级别相同，则打印日志
		if out.Level == -1 || level == out.Level {
			fmt.Fprintln(out.Out, str)
			// 检查并可能处理日志文件的大小
			l.CheckFileSize(out)
		}
	}

}

// WithFields 返回一个新的 Logger 实例，并添加额外的日志字段
func (l *Logger) WithFields(fields Fields) *Logger {
	return &Logger{
		Formatter:    l.Formatter,
		Outs:         l.Outs,
		Level:        l.Level,
		LoggerFields: fields,
	}
}

// SetLogPath 设置日志文件路径，并初始化多个日志文件输出
func (l *Logger) SetLogPath(logPath string) {
	// 设置日志路径并初始化不同级别的日志输出
	// logPath 是日志文件的目录路径
	l.logPath = logPath

	// 添加记录所有级别日志的输出
	l.Outs = append(l.Outs, &LoggerWriter{
		Level: -1, // -1 表示记录所有级别的日志
		Out:   FileWriter(path.Join(logPath, "all.log")),
	})

	// 添加记录调试级别日志的输出
	l.Outs = append(l.Outs, &LoggerWriter{
		Level: LevelDebug, // 仅记录调试级别的日志
		Out:   FileWriter(path.Join(logPath, "debug.log")),
	})

	// 添加记录信息级别日志的输出
	l.Outs = append(l.Outs, &LoggerWriter{
		Level: LevelInfo, // 仅记录信息级别的日志
		Out:   FileWriter(path.Join(logPath, "info.log")),
	})

	// 添加记录错误级别日志的输出
	l.Outs = append(l.Outs, &LoggerWriter{
		Level: LevelError, // 仅记录错误级别的日志
		Out:   FileWriter(path.Join(logPath, "error.log")),
	})
}

// CheckFileSize 检查日志文件大小，如果超过限制则创建新的日志文件
func (l *Logger) CheckFileSize(w *LoggerWriter) {
	// 获取日志文件对象
	logFile := w.Out.(*os.File)
	// 检查日志文件是否有效
	if logFile != nil {
		// 获取日志文件的信息
		stat, err := logFile.Stat()
		if err != nil {
			// 如果获取文件信息失败，则记录错误并退出
			log.Println(err)
			return
		}
		// 获取日志文件的大小
		size := stat.Size()
		// 设置默认的最大日志文件大小为 100MB，如果未设置的话
		if l.LogFileSize <= 0 {
			l.LogFileSize = 100 << 20 // 默认最大日志文件大小为 100MB
		}
		// 检查当前日志文件大小是否超过了最大限制
		if size >= l.LogFileSize {
			// 获取当前日志文件的名称
			_, name := path.Split(stat.Name())
			// 提取文件名，去除扩展名
			fileName := name[0:strings.Index(name, ".")]
			// 创建一个新的日志文件写入对象，文件名基于原始文件名和当前时间戳
			writer := FileWriter(path.Join(l.logPath, string_func.JoinStrings(fileName, ".", time.Now().UnixMilli(), ".log")))
			// 更新输出流为新的日志文件
			w.Out = writer
		}
	}

}

// format 格式化日志消息
func (f *LoggerFormatter) format(msg any) string {
	now := time.Now()
	if f.IsColor {
		levelColor := f.LevelColor()
		msgColor := f.MsgColor()
		return fmt.Sprintf("%s [frame] %s %s%v%s | level= %s %s %s | msg=%s %#v %s | fields=%v ",
			yellow, reset, blue, now.Format("2006/01/02 - 15:04:05"), reset,
			levelColor, f.Level.Level(), reset, msgColor, msg, reset, f.LoggerFields,
		)
	}
	return fmt.Sprintf("[frame] %v | level=%s | msg=%#v | fields=%#v",
		now.Format("2006/01/02 - 15:04:05"),
		f.Level.Level(), msg, f.LoggerFields)
}

// LevelColor 根据日志级别返回相应的颜色代码
func (f *LoggerFormatter) LevelColor() string {
	switch f.Level {
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

// MsgColor 根据日志级别返回消息的颜色代码
func (f *LoggerFormatter) MsgColor() string {
	switch f.Level {
	case LevelError:
		return red
	default:
		return ""
	}
}

// FileWriter 打开或创建一个日志文件，并返回 io.Writer
func FileWriter(name string) io.Writer {
	w, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	return w
}
