package render

import (
	"errors"
	"fmt"
	"net/http"
)

// Redirect 结构体用于定义重定向操作所需的信息。
// 它包括状态码(Code)、请求(Request)和重定向位置(Location)。
type Redirect struct {
	Code     int
	Request  *http.Request
	Location string
}

// Render 方法负责执行实际的HTTP重定向。
// 参数:
//
//	w: http.ResponseWriter，用于向客户端发送响应。
//	code: int，本次重定向使用的HTTP状态码。
//
// 返回值:
//
//	error，如果状态码不适用于重定向，则返回错误。
func (r *Redirect) Render(w http.ResponseWriter, code int) error {
	// 先设置响应的内容类型。
	r.WriteContentType(w)
	// 检查状态码是否适用于重定向，如果不在预期范围内，则返回错误。
	if (r.Code < http.StatusMultipleChoices ||
		r.Code > http.StatusPermanentRedirect) && r.Code != http.StatusCreated {
		return errors.New(fmt.Sprintf("Cannot redirect with status code %d", r.Code))
	}
	// 使用http.Redirect执行实际的重定向操作。
	http.Redirect(w, r.Request, r.Location, r.Code)
	return nil
}

// WriteContentType 方法用于设置响应的内容类型为HTML和UTF-8编码。
// 参数:
//
//	w: http.ResponseWriter，用于向客户端发送响应。
func (r *Redirect) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "text/html; charset=utf-8")
}
