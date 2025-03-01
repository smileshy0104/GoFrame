package frame

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// 定义不同颜色的背景和前景色，用于日志着色
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

// DefaultWriter 是日志的默认输出流
var DefaultWriter io.Writer = os.Stdout

// LoggingConfig 定义了日志的配置，包括格式化函数、输出流和是否着色
type LoggingConfig struct {
	// Formatter 是一个日志格式化函数
	Formatter LoggerFormatter
	// out 是日志的输出流
	out io.Writer
	// IsColor 是一个布尔值，用于指示是否使用着色
	IsColor bool
}

// LoggerFormatter 是一个函数类型，用于格式化日志
type LoggerFormatter = func(params *LogFormatterParams) string

// LogFormatterParams 包含了日志格式化所需的所有参数
type LogFormatterParams struct {
	Request        *http.Request // 请求对象
	TimeStamp      time.Time     // 请求时间戳
	StatusCode     int           // 状态码
	Latency        time.Duration // 延迟时间
	ClientIP       net.IP        // 客户端IP地址
	Method         string        // 请求方法
	Path           string        // 请求路径
	IsDisplayColor bool          // 是否使用着色
}

// StatusCodeColor 根据HTTP状态码返回相应的颜色代码
func (p *LogFormatterParams) StatusCodeColor() string {
	// 根据HTTP状态码返回相应的颜色代码
	code := p.StatusCode
	switch code {
	case http.StatusOK:
		return green
	default:
		return red
	}
}

// ResetColor 返回重置颜色的代码
func (p *LogFormatterParams) ResetColor() string {
	return reset
}

// defaultFormatter 是一个默认的日志格式化函数
var defaultFormatter = func(params *LogFormatterParams) string {
	// 创建一个LogFormatterParams结构体，并设置其属性
	var statusCodeColor = params.StatusCodeColor()
	var resetColor = params.ResetColor()
	if params.Latency > time.Minute {
		params.Latency = params.Latency.Truncate(time.Second)
	}
	// 返回一个格式化后的日志字符串（带颜色）
	if params.IsDisplayColor {
		return fmt.Sprintf("%s [frame] %s |%s %v %s| %s %3d %s |%s %13v %s| %15s  |%s %-7s %s %s %#v %s \n",
			yellow, resetColor, blue, params.TimeStamp.Format("2006/01/02 - 15:04:05"), resetColor,
			statusCodeColor, params.StatusCode, resetColor,
			red, params.Latency, resetColor,
			params.ClientIP,
			magenta, params.Method, resetColor,
			cyan, params.Path, resetColor,
		)
	}
	// 返回一个格式化后的日志字符串（不带颜色）
	return fmt.Sprintf("[frame] %v | %3d | %13v | %15s |%-7s %#v",
		params.TimeStamp.Format("2006/01/02 - 15:04:05"),
		params.StatusCode,
		params.Latency, params.ClientIP, params.Method, params.Path,
	)
}

// LoggingWithConfig 是一个中间件函数，用于根据配置记录HTTP请求的日志。
//
// 参数：
// - conf: LoggingConfig 类型，包含日志的格式化函数、输出流和是否使用颜色等配置项。
// - next: HandlerFunc 类型，表示下一个要执行的处理函数。
//
// 返回值：
// - HandlerFunc 类型，返回一个新的处理函数，该函数在执行完 next 处理函数后会记录请求日志。
func LoggingWithConfig(conf LoggingConfig, next HandlerFunc) HandlerFunc {
	// 如果未提供自定义格式化函数，则使用默认格式化函数
	formatter := conf.Formatter
	if formatter == nil {
		// 使用默认格式化函数
		formatter = defaultFormatter
	}

	// 如果未提供输出流，则使用默认输出流并启用颜色显示
	out := conf.out
	displayColor := false
	if out == nil {
		// 使用默认输出流并启用颜色显示
		out = DefaultWriter
		displayColor = true
	}

	// 返回一个新的处理函数，该函数会在执行完 next 后记录日志
	return func(ctx *Context) {
		r := ctx.R
		// 创建一个LogFormatterParams结构体，并设置其属性
		param := &LogFormatterParams{
			Request:        r,
			IsDisplayColor: displayColor,
		}

		// 记录请求开始时间
		start := time.Now()
		path := r.URL.Path
		raw := r.URL.RawQuery

		// 执行下一个处理函数
		next(ctx)

		// 记录请求结束时间，并计算延迟时间
		stop := time.Now()
		latency := stop.Sub(start)

		// 获取客户端IP地址
		ip, _, _ := net.SplitHostPort(strings.TrimSpace(ctx.R.RemoteAddr))
		clientIP := net.ParseIP(ip)
		method := r.Method
		// 获取状态码
		statusCode := ctx.StatusCode

		// 如果有查询参数，则将其附加到路径中
		if raw != "" {
			path = path + "?" + raw
		}

		// 设置日志格式化参数
		param.TimeStamp = stop
		param.StatusCode = statusCode
		param.Latency = latency
		param.Path = path
		param.ClientIP = clientIP
		param.Method = method

		// 将格式化后的日志输出到指定的输出流
		fmt.Fprint(out, formatter(param))
	}
}

// Logging 是一个中间件，用于使用默认配置记录HTTP请求的日志
func Logging(next HandlerFunc) HandlerFunc {
	// 创建并返回一个中间件函数
	return LoggingWithConfig(LoggingConfig{}, next)
}
