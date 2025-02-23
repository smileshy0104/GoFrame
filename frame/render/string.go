package render

import (
	"fmt"
	"frame/internal/bytesconv"
	"net/http"
)

// String 结构体用于存储格式化字符串及其数据。
// 它提供了一种渲染字符串到 HTTP 响应的方法。
type String struct {
	Format string // 格式化字符串的模板。
	Data   []any  // 格式化字符串所需的数据。
}

// Render 方法将格式化后的字符串渲染到 HTTP 响应中。
// 它首先设置内容类型，然后根据数据填充格式化字符串并输出。
// 如果没有数据，它将直接输出格式化字符串。
// 参数:
//
//	w: HTTP 响应写入器，用于写入内容。
//	code: HTTP 状态码，用于指示响应的状态。
//
// 返回值:
//
//	如果渲染过程中发生错误，返回该错误；否则返回 nil。
func (s *String) Render(w http.ResponseWriter, code int) error {
	// 设置内容类型
	s.WriteContentType(w)
	// 设置 HTTP 响应状态码
	w.WriteHeader(code)
	// 如果存在数据，使用数据格式化字符串并输出
	if len(s.Data) > 0 {
		_, err := fmt.Fprintf(w, s.Format, s.Data...)
		return err
	}
	// 如果没有数据，直接输出格式化字符串
	_, err := w.Write(bytesconv.StringToBytes(s.Format))
	return err
}

// WriteContentType 方法设置 HTTP 响应的内容类型为 "text/plain; charset=utf-8"。
// 参数:
//
//	w: HTTP 响应写入器，用于设置内容类型。
func (s *String) WriteContentType(w http.ResponseWriter) {
	// 调用工具函数设置内容类型
	writeContentType(w, "text/plain; charset=utf-8")
}
