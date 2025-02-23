package render

import "net/http"

// Render 是一个接口，定义了渲染响应的方法（对应声明的接口中的方法都需要实现）
type Render interface {
	// Render 方法用于将内容渲染到 http.ResponseWriter 中
	// 参数 w: 用于写入响应的 http.ResponseWriter
	// 参数 code: HTTP 状态码
	// 返回值 error: 渲染过程中可能发生的错误
	Render(w http.ResponseWriter, code int) error

	// WriteContentType 方法用于设置响应的 Content-Type
	// 参数 w: 用于写入响应的 http.ResponseWriter
	WriteContentType(w http.ResponseWriter)
}

// writeContentType 函数用于设置响应的 Content-Type 头
// 参数 w: 用于写入响应的 http.ResponseWriter
// 参数 value: 要设置的 Content-Type 值
func writeContentType(w http.ResponseWriter, value string) {
	w.Header().Set("Content-type", value)
}
