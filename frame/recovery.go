package frame

import (
	"errors"
	"fmt"
	newerror "frame/error"
	"net/http"
	"runtime"
	"strings"
)

// detailMsg 生成错误的详细信息。
// 参数 err: 捕获到的错误。
// 返回值: 错误的详细信息字符串，包括错误堆栈信息。
func detailMsg(err any) string {
	var sb strings.Builder
	// 获取调用栈信息
	var pcs = make([]uintptr, 32)
	n := runtime.Callers(3, pcs)
	// 将错误信息写入字符串构建器
	sb.WriteString(fmt.Sprintf("%v\n", err))
	// 遍历调用栈信息，获取函数名、文件名和行号
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		// 将调用栈信息写入字符串构建器
		sb.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return sb.String()
}

// Recovery 是一个中间件，用于捕获并处理请求处理过程中的panic。
// 参数 next: 被包装的处理函数。
// 返回值: 包装后的处理函数，能够捕获panic并进行错误处理。
func Recovery(next HandlerFunc) HandlerFunc {
	return func(ctx *Context) {
		defer func() {
			// 捕获panic
			if err := recover(); err != nil {
				// 如果错误实现了MsError接口，则执行MsError的ExecResult方法
				if e := err.(error); e != nil {
					var Error *newerror.MsError
					if errors.As(e, &Error) {
						Error.ExecResult()
						return
					}
				}
				// 记录错误的详细信息
				ctx.Logger.Error(detailMsg(err))
				// 向客户端返回500错误
				ctx.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		// 调用被包装的处理函数
		next(ctx)
	}
}
